package admin

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"runtime"
	"time"

	"better-auth/internal/models"

	"gorm.io/gorm"
)

// AdminService provides admin management functionality
type AdminService struct {
	db *gorm.DB
}

// NewAdminService creates a new admin service
func NewAdminService(db *gorm.DB) *AdminService {
	return &AdminService{db: db}
}

// Initialize implements the PluginService interface
func (s *AdminService) Initialize(db *gorm.DB) error {
	s.db = db
	// Start metrics collection
	go s.startMetricsCollection()
	return nil
}

// Name implements the PluginService interface
func (s *AdminService) Name() string {
	return "admin"
}

// CreateAdmin creates a new admin user
func (s *AdminService) CreateAdmin(
	ctx context.Context,
	userID string,
	role AdminRole,
	createdBy string,
) (*AdminUser, error) {
	// Check if user is already an admin
	var existing AdminUser
	if err := s.db.WithContext(ctx).First(&existing, "user_id = ?", userID).Error; err == nil {
		return nil, fmt.Errorf("user is already an admin")
	}

	admin := &AdminUser{
		ID:          generateID(),
		UserID:      userID,
		Role:        string(role),
		Permissions: DefaultAdminPermissions[role],
		IsSuper:     role == AdminRoleSuper,
		IsActive:    true,
		CreatedBy:   createdBy,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	if err := s.db.WithContext(ctx).Create(admin).Error; err != nil {
		return nil, fmt.Errorf("failed to create admin: %w", err)
	}

	// Log the action
	s.LogAdminAction(ctx, createdBy, AuditActionCreateAdmin, "admin", admin.ID, map[string]any{
		"targetUserId": userID,
		"role":         role,
	}, "")

	return admin, nil
}

// GetAdmin retrieves an admin by user ID
func (s *AdminService) GetAdmin(ctx context.Context, userID string) (*AdminUser, error) {
	var admin AdminUser
	if err := s.db.WithContext(ctx).First(&admin, "user_id = ? AND is_active = true", userID).Error; err != nil {
		return nil, err
	}
	return &admin, nil
}

// GetAdminByID retrieves an admin by admin ID
func (s *AdminService) GetAdminByID(ctx context.Context, adminID string) (*AdminUser, error) {
	var admin AdminUser
	if err := s.db.WithContext(ctx).First(&admin, "id = ? AND is_active = true", adminID).Error; err != nil {
		return nil, err
	}
	return &admin, nil
}

// ListAdmins retrieves all admin users with pagination
func (s *AdminService) ListAdmins(
	ctx context.Context,
	limit, offset int,
) ([]AdminUser, int64, error) {
	var admins []AdminUser
	var total int64

	if err := s.db.WithContext(ctx).Model(&AdminUser{}).Where("is_active = true").Count(&total).Error; err != nil {
		return nil, 0, err
	}

	if err := s.db.WithContext(ctx).
		Where("is_active = true").
		Order("created_at DESC").
		Limit(limit).
		Offset(offset).
		Find(&admins).Error; err != nil {
		return nil, 0, err
	}

	return admins, total, nil
}

// UpdateAdmin updates an admin user
func (s *AdminService) UpdateAdmin(
	ctx context.Context,
	adminID string,
	updates map[string]any,
	updatedBy string,
) error {
	var admin AdminUser
	if err := s.db.WithContext(ctx).First(&admin, "id = ?", adminID).Error; err != nil {
		return fmt.Errorf("admin not found: %w", err)
	}

	if err := s.db.WithContext(ctx).Model(&admin).Updates(updates).Error; err != nil {
		return fmt.Errorf("failed to update admin: %w", err)
	}

	// Log the action
	s.LogAdminAction(ctx, updatedBy, AuditActionUpdateAdmin, "admin", adminID, updates, "")

	return nil
}

// DeleteAdmin soft deletes an admin user
func (s *AdminService) DeleteAdmin(ctx context.Context, adminID string, deletedBy string) error {
	var admin AdminUser
	if err := s.db.WithContext(ctx).First(&admin, "id = ?", adminID).Error; err != nil {
		return fmt.Errorf("admin not found: %w", err)
	}

	if err := s.db.WithContext(ctx).Delete(&admin).Error; err != nil {
		return fmt.Errorf("failed to delete admin: %w", err)
	}

	// Log the action
	s.LogAdminAction(ctx, deletedBy, AuditActionDeleteAdmin, "admin", adminID, map[string]any{
		"targetUserId": admin.UserID,
	}, "")

	return nil
}

// CheckPermission checks if an admin has a specific permission
func (s *AdminService) CheckPermission(
	ctx context.Context,
	userID, permission string,
) (bool, error) {
	admin, err := s.GetAdmin(ctx, userID)
	if err != nil {
		return false, err
	}

	return admin.HasPermission(permission), nil
}

// LogAdminAction logs an admin action for audit purposes
func (s *AdminService) LogAdminAction(
	ctx context.Context,
	adminID, action, resource, resourceID string,
	details map[string]any,
	errorMsg string,
) {
	// Get IP and User Agent from context if available
	ipAddress := ""
	userAgent := ""

	if ip := ctx.Value("ip"); ip != nil {
		if ipStr, ok := ip.(string); ok {
			ipAddress = ipStr
		}
	}

	if ua := ctx.Value("userAgent"); ua != nil {
		if uaStr, ok := ua.(string); ok {
			userAgent = uaStr
		}
	}

	status := "success"
	if errorMsg != "" {
		status = "error"
	}

	auditLog := &AdminAuditLog{
		ID:         generateID(),
		AdminID:    adminID,
		Action:     action,
		Resource:   resource,
		ResourceID: resourceID,
		Details:    details,
		IPAddress:  ipAddress,
		UserAgent:  userAgent,
		Status:     status,
		ErrorMsg:   errorMsg,
		CreatedAt:  time.Now(),
	}

	// Log asynchronously to avoid blocking
	go func() {
		s.db.Create(auditLog)
	}()
}

// GetAuditLogs retrieves audit logs with filtering and pagination
func (s *AdminService) GetAuditLogs(
	ctx context.Context,
	filters map[string]any,
	limit, offset int,
) ([]AdminAuditLog, int64, error) {
	var logs []AdminAuditLog
	var total int64

	query := s.db.WithContext(ctx).Model(&AdminAuditLog{})

	// Apply filters
	if adminID, ok := filters["adminId"]; ok {
		query = query.Where("admin_id = ?", adminID)
	}
	if action, ok := filters["action"]; ok {
		query = query.Where("action = ?", action)
	}
	if resource, ok := filters["resource"]; ok {
		query = query.Where("resource = ?", resource)
	}
	if status, ok := filters["status"]; ok {
		query = query.Where("status = ?", status)
	}
	if startDate, ok := filters["startDate"]; ok {
		query = query.Where("created_at >= ?", startDate)
	}
	if endDate, ok := filters["endDate"]; ok {
		query = query.Where("created_at <= ?", endDate)
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	if err := query.Order("created_at DESC").Limit(limit).Offset(offset).Find(&logs).Error; err != nil {
		return nil, 0, err
	}

	return logs, total, nil
}

// RecordMetric records a system metric
func (s *AdminService) RecordMetric(
	ctx context.Context,
	metricType, metricName string,
	value float64,
	tags []string,
) error {
	metric := &SystemMetrics{
		ID:          generateID(),
		MetricType:  metricType,
		MetricName:  metricName,
		Value:       value,
		Tags:        tags,
		CollectedAt: time.Now(),
		CreatedAt:   time.Now(),
	}

	return s.db.WithContext(ctx).Create(metric).Error
}

// GetMetrics retrieves system metrics with filtering
func (s *AdminService) GetMetrics(
	ctx context.Context,
	metricName string,
	startTime, endTime time.Time,
	tags []string,
) ([]SystemMetrics, error) {
	var metrics []SystemMetrics

	query := s.db.WithContext(ctx).Model(&SystemMetrics{})

	if metricName != "" {
		query = query.Where("metric_name = ?", metricName)
	}

	if !startTime.IsZero() {
		query = query.Where("collected_at >= ?", startTime)
	}

	if !endTime.IsZero() {
		query = query.Where("collected_at <= ?", endTime)
	}

	if len(tags) > 0 {
		for _, tag := range tags {
			query = query.Where("JSON_CONTAINS(tags, ?)", fmt.Sprintf(`"%s"`, tag))
		}
	}

	if err := query.Order("collected_at DESC").Find(&metrics).Error; err != nil {
		return nil, err
	}

	return metrics, nil
}

// GetSystemStats returns system-wide statistics
func (s *AdminService) GetSystemStats(ctx context.Context) (map[string]any, error) {
	stats := make(map[string]any)

	// User statistics
	var userCount int64
	s.db.WithContext(ctx).Model(&models.User{}).Count(&userCount)
	stats["userCount"] = userCount

	var activeUserCount int64
	s.db.WithContext(ctx).Model(&models.User{}).Where("blocked = false").Count(&activeUserCount)
	stats["activeUserCount"] = activeUserCount

	// Session statistics
	var sessionCount int64
	s.db.WithContext(ctx).Model(&models.Session{}).Where("active = true").Count(&sessionCount)
	stats["activeSessionCount"] = sessionCount

	// Admin statistics
	var adminCount int64
	s.db.WithContext(ctx).Model(&AdminUser{}).Where("is_active = true").Count(&adminCount)
	stats["adminCount"] = adminCount

	// Recent activity
	var recentLogins int64
	since := time.Now().Add(-24 * time.Hour)
	s.db.WithContext(ctx).Model(&AdminAuditLog{}).
		Where("action = ? AND created_at >= ?", AuditActionLogin, since).
		Count(&recentLogins)
	stats["recentLogins"] = recentLogins

	return stats, nil
}

// startMetricsCollection starts background metrics collection
func (s *AdminService) startMetricsCollection() {
	ticker := time.NewTicker(5 * time.Minute)
	defer ticker.Stop()

	for range ticker.C {
		ctx := context.Background()

		// Collect system metrics
		var m runtime.MemStats
		runtime.ReadMemStats(&m)

		s.RecordMetric(
			ctx,
			MetricTypeGauge,
			MetricMemoryUsage,
			float64(m.Alloc)/1024/1024,
			[]string{"unit:mb"},
		)

		// Collect user count
		var userCount int64
		s.db.Model(&models.User{}).Count(&userCount)
		s.RecordMetric(ctx, MetricTypeGauge, MetricUserCount, float64(userCount), nil)

		// Collect active session count
		var sessionCount int64
		s.db.Model(&models.Session{}).Where("active = true").Count(&sessionCount)
		s.RecordMetric(ctx, MetricTypeGauge, MetricSessionCount, float64(sessionCount), nil)
	}
}

// Helper functions

func generateID() string {
	bytes := make([]byte, 16)
	rand.Read(bytes)
	return hex.EncodeToString(bytes)
}

// MarshalDetails marshals details map to JSON for database storage
func MarshalDetails(details map[string]any) string {
	if details == nil {
		return ""
	}
	data, _ := json.Marshal(details)
	return string(data)
}

// UnmarshalDetails unmarshals JSON string to details map
func UnmarshalDetails(data string) map[string]any {
	if data == "" {
		return nil
	}
	var details map[string]any
	json.Unmarshal([]byte(data), &details)
	return details
}

