package organizations

import (
	"errors"
	"net/http"
	"strings"
	"time"

	"better-auth/internal/models"
	"better-auth/pkg/transport"
)

// OrganizationHandlers provides HTTP handlers for organization operations
type OrganizationHandlers struct {
	service *OrganizationService
	transport.Transport
}

// NewOrganizationHandlers creates new organization handlers
func NewOrganizationHandlers(service *OrganizationService, transport transport.Transport) *OrganizationHandlers {
	return &OrganizationHandlers{
		service:   service,
		Transport: transport,
	}
}

// CreateOrganization handles organization creation
func (h *OrganizationHandlers) CreateOrganization(w http.ResponseWriter, r *http.Request) {
	var req models.CreateOrganizationRequest
	if err := h.DecodeJSON(r, &req); err != nil {
		var validationErr *transport.ValidationError
		if errors.As(err, &validationErr) {
			h.RespondValidationError(w, validationErr)
		} else {
			h.RespondError(w, http.StatusBadRequest, "Invalid request body")
		}
		return
	}

	// Get user from context (assuming middleware sets this)
	userCtx, ok := h.GetUserContext(r)
	if !ok {
		h.RespondError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	user := userCtx.User

	// Generate slug if not provided
	if req.Slug == "" {
		req.Slug = generateSlug(req.Name)
	}

	org := &Organization{
		ID:        generateID(),
		Name:      req.Name,
		Slug:      req.Slug,
		Logo:      req.Logo,
		Metadata:  req.Metadata,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	if err := h.service.CreateOrganization(r.Context(), org, user.ID); err != nil {
		h.RespondError(w, http.StatusInternalServerError, "Failed to create organization")
		return
	}

	h.RespondJSON(w, http.StatusOK, org)
}

// GetOrganization handles retrieving an organization
func (h *OrganizationHandlers) GetOrganization(w http.ResponseWriter, r *http.Request) {
	orgID := r.PathValue("id")
	if orgID == "" {
		h.RespondError(w, http.StatusBadRequest, "Organization ID required")
		return
	}

	org, err := h.service.GetOrganization(r.Context(), orgID)
	if err != nil {
		h.RespondError(w, http.StatusNotFound, "Organization not found")
		return
	}

	h.RespondJSON(w, http.StatusOK, org)
}

// UpdateOrganization handles organization updates
func (h *OrganizationHandlers) UpdateOrganization(w http.ResponseWriter, r *http.Request) {
	orgID := r.PathValue("id")
	if orgID == "" {
		h.RespondError(w, http.StatusBadRequest, "Organization ID required")
		return
	}

	var req models.CreateOrganizationRequest
	if err := h.DecodeJSON(r, &req); err != nil {
		var validationErr *transport.ValidationError
		if errors.As(err, &validationErr) {
			h.RespondValidationError(w, validationErr)
		} else {
			h.RespondError(w, http.StatusBadRequest, "Invalid request body")
		}
		return
	}

	org, err := h.service.GetOrganization(r.Context(), orgID)
	if err != nil {
		h.RespondError(w, http.StatusNotFound, "Organization not found")
		return
	}

	// Update fields
	if req.Name != "" {
		org.Name = req.Name
	}
	if req.Slug != "" {
		org.Slug = req.Slug
	}
	if req.Logo != "" {
		org.Logo = req.Logo
	}
	if req.Metadata != nil {
		org.Metadata = req.Metadata
	}
	org.UpdatedAt = time.Now()

	if err := h.service.UpdateOrganization(r.Context(), org); err != nil {
		h.RespondError(w, http.StatusInternalServerError, "Failed to update organization")
		return
	}

	h.RespondJSON(w, http.StatusOK, org)
}

// DeleteOrganization handles organization deletion
func (h *OrganizationHandlers) DeleteOrganization(w http.ResponseWriter, r *http.Request) {
	orgID := r.PathValue("id")
	if orgID == "" {
		h.RespondError(w, http.StatusBadRequest, "Organization ID required")
		return
	}

	if err := h.service.DeleteOrganization(r.Context(), orgID); err != nil {
		h.RespondError(w, http.StatusInternalServerError, "Failed to delete organization")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// InviteToOrganization handles organization invitations
func (h *OrganizationHandlers) InviteToOrganization(w http.ResponseWriter, r *http.Request) {
	orgID := r.PathValue("id")
	if orgID == "" {
		h.RespondError(w, http.StatusBadRequest, "Organization ID required")
		return
	}

	var req models.InviteToOrganizationRequest
	if err := h.DecodeJSON(r, &req); err != nil {
		var validationErr *transport.ValidationError
		if errors.As(err, &validationErr) {
			h.RespondValidationError(w, validationErr)
		} else {
			h.RespondError(w, http.StatusBadRequest, "Invalid request body")
		}
		return
	}

	// Get user from context
	userCtx, ok := h.GetUserContext(r)
	if !ok {
		h.RespondError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	user := userCtx.User

	invitation := &OrganizationInvitation{
		ID:             generateID(),
		OrganizationID: orgID,
		Email:          req.Email,
		Role:           req.Role,
		Status:         string(InvitationPending),
		InvitedBy:      user.ID,
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
	}

	if err := h.service.InviteToOrganization(r.Context(), invitation); err != nil {
		h.RespondError(w, http.StatusInternalServerError, "Failed to create invitation")
		return
	}

	h.RespondJSON(w, http.StatusOK, map[string]string{
		"message": "Invitation sent successfully",
	})
}

// AcceptInvitation handles accepting organization invitations
func (h *OrganizationHandlers) AcceptInvitation(w http.ResponseWriter, r *http.Request) {
	token := r.URL.Query().Get("token")
	if token == "" {
		h.RespondError(w, http.StatusBadRequest, "Invitation token required")
		return
	}

	// Get user from context
	userCtx, ok := h.GetUserContext(r)
	if !ok {
		h.RespondError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	user := userCtx.User

	if err := h.service.AcceptInvitation(r.Context(), token, user.ID); err != nil {
		h.RespondError(w, http.StatusBadRequest, "Failed to accept invitation")
		return
	}

	h.RespondJSON(w, http.StatusOK, map[string]string{
		"message": "Invitation accepted successfully",
	})
}

// GetOrganizationMembers handles retrieving organization members
func (h *OrganizationHandlers) GetOrganizationMembers(w http.ResponseWriter, r *http.Request) {
	orgID := r.PathValue("id")
	if orgID == "" {
		h.RespondError(w, http.StatusBadRequest, "Organization ID required")
		return
	}

	members, err := h.service.GetOrganizationMembers(r.Context(), orgID)
	if err != nil {
		h.RespondError(w, http.StatusInternalServerError, "Failed to get members")
		return
	}

	h.RespondJSON(w, http.StatusOK, members)
}

// UpdateMemberRole handles updating a member's role
func (h *OrganizationHandlers) UpdateMemberRole(w http.ResponseWriter, r *http.Request) {
	memberID := r.PathValue("memberId")
	if memberID == "" {
		h.RespondError(w, http.StatusBadRequest, "Member ID required")
		return
	}

	var req models.UpdateMemberRoleRequest
	if err := h.DecodeJSON(r, &req); err != nil {
		var validationErr *transport.ValidationError
		if errors.As(err, &validationErr) {
			h.RespondValidationError(w, validationErr)
		} else {
			h.RespondError(w, http.StatusBadRequest, "Invalid request body")
		}
		return
	}

	if err := h.service.UpdateMemberRole(r.Context(), memberID, req.Role); err != nil {
		h.RespondError(w, http.StatusInternalServerError, "Failed to update member role")
		return
	}

	h.RespondJSON(w, http.StatusOK, map[string]string{
		"message": "Member role updated successfully",
	})
}

// RemoveMember handles removing a member from an organization
func (h *OrganizationHandlers) RemoveMember(w http.ResponseWriter, r *http.Request) {
	memberID := r.PathValue("memberId")
	if memberID == "" {
		h.RespondError(w, http.StatusBadRequest, "Member ID required")
		return
	}

	if err := h.service.RemoveMember(r.Context(), memberID); err != nil {
		h.RespondError(w, http.StatusInternalServerError, "Failed to remove member")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// CreateTeam handles team creation
func (h *OrganizationHandlers) CreateTeam(w http.ResponseWriter, r *http.Request) {
	orgID := r.PathValue("id")
	if orgID == "" {
		h.RespondError(w, http.StatusBadRequest, "Organization ID required")
		return
	}

	var req struct {
		Name        string `json:"name"`
		Description string `json:"description,omitempty"`
		Color       string `json:"color,omitempty"`
	}

	if err := h.DecodeJSON(r, &req); err != nil {
		var validationErr *transport.ValidationError
		if errors.As(err, &validationErr) {
			h.RespondValidationError(w, validationErr)
		} else {
			h.RespondError(w, http.StatusBadRequest, "Invalid request body")
		}
		return
	}

	team := &Team{
		ID:             generateID(),
		OrganizationID: orgID,
		Name:           req.Name,
		Slug:           generateSlug(req.Name),
		Description:    req.Description,
		Color:          req.Color,
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
	}

	if err := h.service.CreateTeam(r.Context(), team); err != nil {
		h.RespondError(w, http.StatusInternalServerError, "Failed to create team")
		return
	}

	h.RespondJSON(w, http.StatusOK, team)
}

// GetOrganizationTeams handles retrieving organization teams
func (h *OrganizationHandlers) GetOrganizationTeams(w http.ResponseWriter, r *http.Request) {
	orgID := r.PathValue("id")
	if orgID == "" {
		h.RespondError(w, http.StatusBadRequest, "Organization ID required")
		return
	}

	teams, err := h.service.GetOrganizationTeams(r.Context(), orgID)
	if err != nil {
		h.RespondError(w, http.StatusInternalServerError, "Failed to get teams")
		return
	}

	h.RespondJSON(w, http.StatusOK, teams)
}

// Helper function to generate slug from name
func generateSlug(name string) string {
	slug := strings.ToLower(name)
	slug = strings.ReplaceAll(slug, " ", "-")
	// Remove special characters
	result := ""
	for _, r := range slug {
		if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') || r == '-' {
			result += string(r)
		}
	}
	return result
}
