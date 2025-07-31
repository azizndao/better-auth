package admin

import (
	"errors"
	"net/http"
	"strconv"
	"time"

	"better-auth/pkg/transport"
)

// AdminHandlers provides HTTP handlers for admin operations
type AdminHandlers struct {
	service *AdminService
	transport.Transport
}

// NewAdminHandlers creates new admin handlers
func NewAdminHandlers(service *AdminService, transport transport.Transport) *AdminHandlers {
	return &AdminHandlers{
		service:   service,
		Transport: transport,
	}
}

// CreateAdminRequest represents a request to create a new admin
type CreateAdminRequest struct {
	UserID string    `json:"userId" validate:"required"`
	Role   AdminRole `json:"role"   validate:"required"`
}

// UpdateAdminRequest represents a request to update an admin
type UpdateAdminRequest struct {
	Role        string   `json:"role,omitempty"`
	Permissions []string `json:"permissions,omitempty"`
	IsActive    *bool    `json:"isActive,omitempty"`
}

// CreateAdmin handles creating a new admin user
func (h *AdminHandlers) CreateAdmin(w http.ResponseWriter, r *http.Request) {
	// Check admin permissions
	if !h.checkPermission(r, PermissionManageAdmins) {
		h.RespondError(w, http.StatusForbidden, "Insufficient permissions")
		return
	}

	var req CreateAdminRequest
	if err := h.DecodeJSON(r, &req); err != nil {
		var validationErr *transport.ValidationError
		if errors.As(err, &validationErr) {
			h.RespondValidationError(w, validationErr)
		} else {
			h.RespondError(w, http.StatusBadRequest, "Invalid request body")
		}
		return
	}

	// Get current admin user
	currentAdmin := h.getCurrentAdmin(r)
	if currentAdmin == nil {
		h.RespondError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	admin, err := h.service.CreateAdmin(r.Context(), req.UserID, req.Role, currentAdmin.ID)
	if err != nil {
		h.RespondError(w, http.StatusInternalServerError, "Failed to create admin: "+err.Error())
		return
	}

	h.RespondJSON(w, http.StatusOK, admin)
}

// GetAdmin handles retrieving admin information
func (h *AdminHandlers) GetAdmin(w http.ResponseWriter, r *http.Request) {
	if !h.checkPermission(r, PermissionViewAdmins) {
		h.RespondError(w, http.StatusForbidden, "Insufficient permissions")
		return
	}

	adminID := r.PathValue("id")
	if adminID == "" {
		h.RespondError(w, http.StatusBadRequest, "Admin ID required")
		return
	}

	admin, err := h.service.GetAdminByID(r.Context(), adminID)
	if err != nil {
		h.RespondError(w, http.StatusNotFound, "Admin not found")
		return
	}

	h.RespondJSON(w, http.StatusOK, admin)
}

// ListAdmins handles listing all admin users
func (h *AdminHandlers) ListAdmins(w http.ResponseWriter, r *http.Request) {
	if !h.checkPermission(r, PermissionViewAdmins) {
		h.RespondError(w, http.StatusForbidden, "Insufficient permissions")
		return
	}

	// Parse pagination parameters
	limitStr := r.URL.Query().Get("limit")
	offsetStr := r.URL.Query().Get("offset")

	limit := 20 // default
	offset := 0 // default

	if limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 {
			limit = l
		}
	}

	if offsetStr != "" {
		if o, err := strconv.Atoi(offsetStr); err == nil && o >= 0 {
			offset = o
		}
	}

	admins, total, err := h.service.ListAdmins(r.Context(), limit, offset)
	if err != nil {
		h.RespondError(w, http.StatusInternalServerError, "Failed to retrieve admins")
		return
	}

	response := map[string]any{
		"admins": admins,
		"total":  total,
		"limit":  limit,
		"offset": offset,
	}

	h.RespondJSON(w, http.StatusOK, response)
}

// UpdateAdmin handles updating an admin user
func (h *AdminHandlers) UpdateAdmin(w http.ResponseWriter, r *http.Request) {
	if !h.checkPermission(r, PermissionManageAdmins) {
		h.RespondError(w, http.StatusForbidden, "Insufficient permissions")
		return
	}

	adminID := r.PathValue("id")
	if adminID == "" {
		h.RespondError(w, http.StatusBadRequest, "Admin ID required")
		return
	}

	var req UpdateAdminRequest
	if err := h.DecodeJSON(r, &req); err != nil {
		var validationErr *transport.ValidationError
		if errors.As(err, &validationErr) {
			h.RespondValidationError(w, validationErr)
		} else {
			h.RespondError(w, http.StatusBadRequest, "Invalid request body")
		}
		return
	}

	// Build updates map
	updates := make(map[string]any)
	if req.Role != "" {
		updates["role"] = req.Role
		// Update permissions based on role
		if role := AdminRole(req.Role); DefaultAdminPermissions[role] != nil {
			updates["permissions"] = DefaultAdminPermissions[role]
		}
	}
	if req.Permissions != nil {
		updates["permissions"] = req.Permissions
	}
	if req.IsActive != nil {
		updates["is_active"] = *req.IsActive
	}
	updates["updated_at"] = time.Now()

	currentAdmin := h.getCurrentAdmin(r)
	if currentAdmin == nil {
		h.RespondError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	if err := h.service.UpdateAdmin(r.Context(), adminID, updates, currentAdmin.ID); err != nil {
		h.RespondError(w, http.StatusInternalServerError, "Failed to update admin: "+err.Error())
		return
	}

	h.RespondJSON(w, http.StatusOK, map[string]string{
		"message": "Admin updated successfully",
	})
}

// DeleteAdmin handles deleting an admin user
func (h *AdminHandlers) DeleteAdmin(w http.ResponseWriter, r *http.Request) {
	if !h.checkPermission(r, PermissionManageAdmins) {
		h.RespondError(w, http.StatusForbidden, "Insufficient permissions")
		return
	}

	adminID := r.PathValue("id")
	if adminID == "" {
		h.RespondError(w, http.StatusBadRequest, "Admin ID required")
		return
	}

	currentAdmin := h.getCurrentAdmin(r)
	if currentAdmin == nil {
		h.RespondError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	// Prevent self-deletion
	if adminID == currentAdmin.ID {
		h.RespondError(w, http.StatusBadRequest, "Cannot delete your own admin account")
		return
	}

	if err := h.service.DeleteAdmin(r.Context(), adminID, currentAdmin.ID); err != nil {
		h.RespondError(w, http.StatusInternalServerError, "Failed to delete admin: "+err.Error())
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// GetAuditLogs handles retrieving audit logs
func (h *AdminHandlers) GetAuditLogs(w http.ResponseWriter, r *http.Request) {
	if !h.checkPermission(r, PermissionViewAuditLogs) {
		h.RespondError(w, http.StatusForbidden, "Insufficient permissions")
		return
	}

	// Parse query parameters
	query := r.URL.Query()
	filters := make(map[string]any)

	if adminID := query.Get("adminId"); adminID != "" {
		filters["adminId"] = adminID
	}
	if action := query.Get("action"); action != "" {
		filters["action"] = action
	}
	if resource := query.Get("resource"); resource != "" {
		filters["resource"] = resource
	}
	if status := query.Get("status"); status != "" {
		filters["status"] = status
	}
	if startDate := query.Get("startDate"); startDate != "" {
		if t, err := time.Parse(time.RFC3339, startDate); err == nil {
			filters["startDate"] = t
		}
	}
	if endDate := query.Get("endDate"); endDate != "" {
		if t, err := time.Parse(time.RFC3339, endDate); err == nil {
			filters["endDate"] = t
		}
	}

	// Parse pagination
	limit := 50
	offset := 0
	if l := query.Get("limit"); l != "" {
		if parsed, err := strconv.Atoi(l); err == nil && parsed > 0 {
			limit = parsed
		}
	}
	if o := query.Get("offset"); o != "" {
		if parsed, err := strconv.Atoi(o); err == nil && parsed >= 0 {
			offset = parsed
		}
	}

	logs, total, err := h.service.GetAuditLogs(r.Context(), filters, limit, offset)
	if err != nil {
		h.RespondError(w, http.StatusInternalServerError, "Failed to retrieve audit logs")
		return
	}

	response := map[string]any{
		"logs":   logs,
		"total":  total,
		"limit":  limit,
		"offset": offset,
	}

	h.RespondJSON(w, http.StatusOK, response)
}

// GetSystemStats handles retrieving system statistics
func (h *AdminHandlers) GetSystemStats(w http.ResponseWriter, r *http.Request) {
	if !h.checkPermission(r, PermissionViewSystem) {
		h.RespondError(w, http.StatusForbidden, "Insufficient permissions")
		return
	}

	stats, err := h.service.GetSystemStats(r.Context())
	if err != nil {
		h.RespondError(w, http.StatusInternalServerError, "Failed to retrieve system stats")
		return
	}

	h.RespondJSON(w, http.StatusOK, stats)
}

// GetMetrics handles retrieving system metrics
func (h *AdminHandlers) GetMetrics(w http.ResponseWriter, r *http.Request) {
	if !h.checkPermission(r, PermissionViewMetrics) {
		h.RespondError(w, http.StatusForbidden, "Insufficient permissions")
		return
	}

	query := r.URL.Query()
	metricName := query.Get("metric")

	var startTime, endTime time.Time
	if start := query.Get("startTime"); start != "" {
		if t, err := time.Parse(time.RFC3339, start); err == nil {
			startTime = t
		}
	}
	if end := query.Get("endTime"); end != "" {
		if t, err := time.Parse(time.RFC3339, end); err == nil {
			endTime = t
		}
	}

	tags := query["tags"]

	metrics, err := h.service.GetMetrics(r.Context(), metricName, startTime, endTime, tags)
	if err != nil {
		h.RespondError(w, http.StatusInternalServerError, "Failed to retrieve metrics")
		return
	}

	h.RespondJSON(w, http.StatusOK, metrics)
}

// Helper methods

func (h *AdminHandlers) getCurrentAdmin(r *http.Request) *AdminUser {
	userCtx, ok := h.GetUserContext(r)
	if !ok {
		return nil
	}

	user := userCtx.User
	admin, err := h.service.GetAdmin(r.Context(), user.ID)
	if err != nil {
		return nil
	}

	return admin
}

func (h *AdminHandlers) checkPermission(r *http.Request, permission string) bool {
	admin := h.getCurrentAdmin(r)
	if admin == nil {
		return false
	}

	return admin.HasPermission(permission)
}
