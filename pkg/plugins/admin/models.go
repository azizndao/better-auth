package admin

import (
	"slices"
	"time"

	"gorm.io/gorm"
)

// AdminUser represents an admin user with enhanced privileges
type AdminUser struct {
	ID          string         `gorm:"primaryKey"               json:"id"`
	UserID      string         `gorm:"uniqueIndex;not null"     json:"userId"`
	Role        string         `gorm:"not null;default:'admin'" json:"role"`
	Permissions []string       `gorm:"serializer:json"          json:"permissions"`
	IsSuper     bool           `gorm:"default:false"            json:"isSuper"`
	IsActive    bool           `gorm:"default:true"             json:"isActive"`
	CreatedBy   string         `                                json:"createdBy,omitempty"`
	CreatedAt   time.Time      `                                json:"createdAt"`
	UpdatedAt   time.Time      `                                json:"updatedAt"`
	DeletedAt   gorm.DeletedAt `gorm:"index"                    json:"-"`
}

// AdminSession tracks admin login sessions
type AdminSession struct {
	ID        string         `gorm:"primaryKey"     json:"id"`
	AdminID   string         `gorm:"not null;index" json:"adminId"`
	IPAddress string         `                      json:"ipAddress"`
	UserAgent string         `                      json:"userAgent"`
	LoginAt   time.Time      `                      json:"loginAt"`
	LogoutAt  *time.Time     `                      json:"logoutAt,omitempty"`
	IsActive  bool           `gorm:"default:true"   json:"isActive"`
	CreatedAt time.Time      `                      json:"createdAt"`
	UpdatedAt time.Time      `                      json:"updatedAt"`
	DeletedAt gorm.DeletedAt `gorm:"index"          json:"-"`
}

// AdminAuditLog tracks admin actions for security and compliance
type AdminAuditLog struct {
	ID         string         `gorm:"primaryKey"        json:"id"`
	AdminID    string         `gorm:"not null;index"    json:"adminId"`
	Action     string         `gorm:"not null"          json:"action"`
	Resource   string         `gorm:"not null"          json:"resource"`
	ResourceID string         `                         json:"resourceId,omitempty"`
	Details    map[string]any `gorm:"serializer:json"   json:"details,omitempty"`
	IPAddress  string         `                         json:"ipAddress"`
	UserAgent  string         `                         json:"userAgent"`
	Status     string         `gorm:"default:'success'" json:"status"`
	ErrorMsg   string         `                         json:"errorMsg,omitempty"`
	CreatedAt  time.Time      `                         json:"createdAt"`
}

// SystemMetrics stores system-wide metrics and statistics
type SystemMetrics struct {
	ID          string    `gorm:"primaryKey"      json:"id"`
	MetricType  string    `gorm:"not null;index"  json:"metricType"`
	MetricName  string    `gorm:"not null"        json:"metricName"`
	Value       float64   `                       json:"value"`
	Tags        []string  `gorm:"serializer:json" json:"tags,omitempty"`
	CollectedAt time.Time `gorm:"index"           json:"collectedAt"`
	CreatedAt   time.Time `                       json:"createdAt"`
}

// AdminRole defines the available admin roles
type AdminRole string

const (
	AdminRoleAdmin     AdminRole = "admin"
	AdminRoleSuper     AdminRole = "super"
	AdminRoleModerator AdminRole = "moderator"
	AdminRoleSupport   AdminRole = "support"
	AdminRoleAnalyst   AdminRole = "analyst"
)

// AdminPermission defines admin permissions
const (
	PermissionManageUsers      = "users:manage"
	PermissionViewUsers        = "users:view"
	PermissionDeleteUsers      = "users:delete"
	PermissionManageOrgs       = "orgs:manage"
	PermissionViewOrgs         = "orgs:view"
	PermissionDeleteOrgs       = "orgs:delete"
	PermissionManageAdmins     = "admins:manage"
	PermissionViewAdmins       = "admins:view"
	PermissionManageSystem     = "system:manage"
	PermissionViewSystem       = "system:view"
	PermissionViewMetrics      = "metrics:view"
	PermissionManageMetrics    = "metrics:manage"
	PermissionViewAuditLogs    = "audit:view"
	PermissionManageSettings   = "settings:manage"
	PermissionViewSettings     = "settings:view"
	PermissionImpersonateUsers = "users:impersonate"
	PermissionBulkOperations   = "bulk:operations"
)

// DefaultAdminPermissions maps roles to their default permissions
var DefaultAdminPermissions = map[AdminRole][]string{
	AdminRoleSuper: {
		PermissionManageUsers,
		PermissionViewUsers,
		PermissionDeleteUsers,
		PermissionManageOrgs,
		PermissionViewOrgs,
		PermissionDeleteOrgs,
		PermissionManageAdmins,
		PermissionViewAdmins,
		PermissionManageSystem,
		PermissionViewSystem,
		PermissionViewMetrics,
		PermissionManageMetrics,
		PermissionViewAuditLogs,
		PermissionManageSettings,
		PermissionViewSettings,
		PermissionImpersonateUsers,
		PermissionBulkOperations,
	},
	AdminRoleAdmin: {
		PermissionManageUsers,
		PermissionViewUsers,
		PermissionManageOrgs,
		PermissionViewOrgs,
		PermissionViewAdmins,
		PermissionViewSystem,
		PermissionViewMetrics,
		PermissionViewAuditLogs,
		PermissionViewSettings,
		PermissionBulkOperations,
	},
	AdminRoleModerator: {
		PermissionViewUsers,
		PermissionViewOrgs,
		PermissionViewSystem,
		PermissionViewMetrics,
		PermissionViewAuditLogs,
	},
	AdminRoleSupport: {
		PermissionViewUsers,
		PermissionViewOrgs,
		PermissionViewSystem,
		PermissionViewMetrics,
		PermissionImpersonateUsers,
	},
	AdminRoleAnalyst: {
		PermissionViewUsers,
		PermissionViewOrgs,
		PermissionViewSystem,
		PermissionViewMetrics,
		PermissionManageMetrics,
		PermissionViewAuditLogs,
	},
}

// AuditAction defines types of actions that can be audited
const (
	AuditActionLogin           = "login"
	AuditActionLogout          = "logout"
	AuditActionCreateUser      = "create_user"
	AuditActionUpdateUser      = "update_user"
	AuditActionDeleteUser      = "delete_user"
	AuditActionCreateOrg       = "create_org"
	AuditActionUpdateOrg       = "update_org"
	AuditActionDeleteOrg       = "delete_org"
	AuditActionCreateAdmin     = "create_admin"
	AuditActionUpdateAdmin     = "update_admin"
	AuditActionDeleteAdmin     = "delete_admin"
	AuditActionUpdateSettings  = "update_settings"
	AuditActionImpersonateUser = "impersonate_user"
	AuditActionBulkOperation   = "bulk_operation"
	AuditActionSystemCommand   = "system_command"
)

// MetricType defines types of system metrics
const (
	MetricTypeCounter   = "counter"
	MetricTypeGauge     = "gauge"
	MetricTypeHistogram = "histogram"
	MetricTypeSummary   = "summary"
)

// Predefined metric names
const (
	MetricUserCount           = "user_count"
	MetricActiveUserCount     = "active_user_count"
	MetricOrganizationCount   = "organization_count"
	MetricSessionCount        = "session_count"
	MetricLoginCount          = "login_count"
	MetricFailedLoginCount    = "failed_login_count"
	MetricAPIRequestCount     = "api_request_count"
	MetricAPIResponseTime     = "api_response_time"
	MetricDatabaseConnections = "database_connections"
	MetricMemoryUsage         = "memory_usage"
	MetricCPUUsage            = "cpu_usage"
)

// TableName overrides the table name for AdminUser
func (AdminUser) TableName() string {
	return "admin_users"
}

// TableName overrides the table name for AdminSession
func (AdminSession) TableName() string {
	return "admin_sessions"
}

// TableName overrides the table name for AdminAuditLog
func (AdminAuditLog) TableName() string {
	return "admin_audit_logs"
}

// TableName overrides the table name for SystemMetrics
func (SystemMetrics) TableName() string {
	return "system_metrics"
}

// HasPermission checks if the admin user has a specific permission
func (a *AdminUser) HasPermission(permission string) bool {
	if a.IsSuper {
		return true
	}

	return slices.Contains(a.Permissions, permission)
}

// CanAccess checks if the admin user can access a resource
func (a *AdminUser) CanAccess(resource, action string) bool {
	if a.IsSuper {
		return true
	}

	permission := resource + ":" + action
	return a.HasPermission(permission)
}

