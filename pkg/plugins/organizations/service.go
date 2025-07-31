package organizations

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"slices"
	"time"

	"gorm.io/gorm"
)

// OrganizationService provides organization management functionality
type OrganizationService struct {
	db *gorm.DB
}

// NewOrganizationService creates a new organization service
func NewOrganizationService(db *gorm.DB) *OrganizationService {
	return &OrganizationService{db: db}
}

// Initialize implements the PluginService interface
func (s *OrganizationService) Initialize(db *gorm.DB) error {
	s.db = db
	return nil
}

// Name implements the PluginService interface
func (s *OrganizationService) Name() string {
	return "organization"
}

// CreateOrganization creates a new organization
func (s *OrganizationService) CreateOrganization(
	ctx context.Context,
	org *Organization,
	ownerUserID string,
) error {
	return s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// Create organization
		if err := tx.Create(org).Error; err != nil {
			return fmt.Errorf("failed to create organization: %w", err)
		}

		// Create owner membership
		member := &OrganizationMember{
			ID:             generateID(),
			OrganizationID: org.ID,
			UserID:         ownerUserID,
			Role:           string(RoleOwner),
			Status:         string(MemberActive),
			Permissions:    DefaultPermissions[RoleOwner],
			JoinedAt:       time.Now(),
		}

		if err := tx.Create(member).Error; err != nil {
			return fmt.Errorf("failed to create owner membership: %w", err)
		}

		// Create default team
		defaultTeam := &Team{
			ID:             generateID(),
			OrganizationID: org.ID,
			Name:           "Default",
			Slug:           "default",
			Description:    "Default team for " + org.Name,
			IsDefault:      true,
		}

		if err := tx.Create(defaultTeam).Error; err != nil {
			return fmt.Errorf("failed to create default team: %w", err)
		}

		// Add owner to default team
		teamMember := &TeamMember{
			ID:       generateID(),
			TeamID:   defaultTeam.ID,
			UserID:   ownerUserID,
			Role:     string(TeamRoleManager),
			JoinedAt: time.Now(),
		}

		if err := tx.Create(teamMember).Error; err != nil {
			return fmt.Errorf("failed to add owner to default team: %w", err)
		}

		return nil
	})
}

// GetOrganization retrieves an organization by ID
func (s *OrganizationService) GetOrganization(
	ctx context.Context,
	id string,
) (*Organization, error) {
	var org Organization
	if err := s.db.WithContext(ctx).
		Preload("Members").
		Preload("Teams").
		First(&org, "id = ?", id).Error; err != nil {
		return nil, err
	}
	return &org, nil
}

// GetOrganizationBySlug retrieves an organization by slug
func (s *OrganizationService) GetOrganizationBySlug(
	ctx context.Context,
	slug string,
) (*Organization, error) {
	var org Organization
	if err := s.db.WithContext(ctx).
		Preload("Members").
		Preload("Teams").
		First(&org, "slug = ?", slug).Error; err != nil {
		return nil, err
	}
	return &org, nil
}

// UpdateOrganization updates an organization
func (s *OrganizationService) UpdateOrganization(ctx context.Context, org *Organization) error {
	return s.db.WithContext(ctx).Save(org).Error
}

// DeleteOrganization soft deletes an organization
func (s *OrganizationService) DeleteOrganization(ctx context.Context, id string) error {
	return s.db.WithContext(ctx).Delete(&Organization{}, "id = ?", id).Error
}

// GetUserOrganizations retrieves all organizations for a user
func (s *OrganizationService) GetUserOrganizations(
	ctx context.Context,
	userID string,
) ([]Organization, error) {
	var orgs []Organization
	if err := s.db.WithContext(ctx).
		Joins("JOIN organization_members ON organizations.id = organization_members.organization_id").
		Where("organization_members.user_id = ? AND organization_members.status = ?", userID, MemberActive).
		Find(&orgs).Error; err != nil {
		return nil, err
	}
	return orgs, nil
}

// InviteToOrganization creates an invitation to join an organization
func (s *OrganizationService) InviteToOrganization(
	ctx context.Context,
	invitation *OrganizationInvitation,
) error {
	// Generate invitation token
	token, err := generateToken()
	if err != nil {
		return fmt.Errorf("failed to generate invitation token: %w", err)
	}
	invitation.Token = token

	// Set expiration (7 days from now)
	invitation.ExpiresAt = time.Now().Add(7 * 24 * time.Hour)

	return s.db.WithContext(ctx).Create(invitation).Error
}

// AcceptInvitation accepts an organization invitation
func (s *OrganizationService) AcceptInvitation(
	ctx context.Context,
	token string,
	userID string,
) error {
	return s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// Find invitation
		var invitation OrganizationInvitation
		if err := tx.First(&invitation, "token = ? AND status = ?", token, InvitationPending).Error; err != nil {
			return fmt.Errorf("invitation not found: %w", err)
		}

		// Check if invitation is expired
		if time.Now().After(invitation.ExpiresAt) {
			invitation.Status = string(InvitationExpired)
			tx.Save(&invitation)
			return fmt.Errorf("invitation has expired")
		}

		// Create member
		member := &OrganizationMember{
			ID:             generateID(),
			OrganizationID: invitation.OrganizationID,
			UserID:         userID,
			Role:           invitation.Role,
			Status:         string(MemberActive),
			Permissions:    DefaultPermissions[OrganizationRole(invitation.Role)],
			JoinedAt:       time.Now(),
		}

		if err := tx.Create(member).Error; err != nil {
			return fmt.Errorf("failed to create member: %w", err)
		}

		// Update invitation
		now := time.Now()
		invitation.Status = string(InvitationAccepted)
		invitation.AcceptedAt = &now
		if err := tx.Save(&invitation).Error; err != nil {
			return fmt.Errorf("failed to update invitation: %w", err)
		}

		return nil
	})
}

// GetOrganizationMembers retrieves all members of an organization
func (s *OrganizationService) GetOrganizationMembers(
	ctx context.Context,
	orgID string,
) ([]OrganizationMember, error) {
	var members []OrganizationMember
	if err := s.db.WithContext(ctx).
		Where("organization_id = ? AND status = ?", orgID, MemberActive).
		Find(&members).Error; err != nil {
		return nil, err
	}
	return members, nil
}

// UpdateMemberRole updates a member's role
func (s *OrganizationService) UpdateMemberRole(
	ctx context.Context,
	memberID string,
	role string,
) error {
	member := &OrganizationMember{}
	if err := s.db.WithContext(ctx).First(member, "id = ?", memberID).Error; err != nil {
		return err
	}

	member.Role = role
	member.Permissions = DefaultPermissions[OrganizationRole(role)]
	return s.db.WithContext(ctx).Save(member).Error
}

// RemoveMember removes a member from an organization
func (s *OrganizationService) RemoveMember(ctx context.Context, memberID string) error {
	return s.db.WithContext(ctx).Delete(&OrganizationMember{}, "id = ?", memberID).Error
}

// CreateTeam creates a new team within an organization
func (s *OrganizationService) CreateTeam(ctx context.Context, team *Team) error {
	return s.db.WithContext(ctx).Create(team).Error
}

// GetTeam retrieves a team by ID
func (s *OrganizationService) GetTeam(ctx context.Context, id string) (*Team, error) {
	var team Team
	if err := s.db.WithContext(ctx).
		Preload("Members").
		First(&team, "id = ?", id).Error; err != nil {
		return nil, err
	}
	return &team, nil
}

// GetOrganizationTeams retrieves all teams for an organization
func (s *OrganizationService) GetOrganizationTeams(
	ctx context.Context,
	orgID string,
) ([]Team, error) {
	var teams []Team
	if err := s.db.WithContext(ctx).
		Where("organization_id = ?", orgID).
		Find(&teams).Error; err != nil {
		return nil, err
	}
	return teams, nil
}

// AddTeamMember adds a user to a team
func (s *OrganizationService) AddTeamMember(ctx context.Context, teamMember *TeamMember) error {
	return s.db.WithContext(ctx).Create(teamMember).Error
}

// RemoveTeamMember removes a user from a team
func (s *OrganizationService) RemoveTeamMember(ctx context.Context, teamID, userID string) error {
	return s.db.WithContext(ctx).
		Delete(&TeamMember{}, "team_id = ? AND user_id = ?", teamID, userID).
		Error
}

// CheckPermission checks if a user has a specific permission in an organization
func (s *OrganizationService) CheckPermission(
	ctx context.Context,
	userID, orgID, permission string,
) (bool, error) {
	var member OrganizationMember
	if err := s.db.WithContext(ctx).
		First(&member, "user_id = ? AND organization_id = ? AND status = ?", userID, orgID, MemberActive).Error; err != nil {
		return false, err
	}

	if slices.Contains(member.Permissions, permission) {
		return true, nil
	}
	return false, nil
}

// Helper functions

func generateID() string {
	bytes := make([]byte, 16)
	rand.Read(bytes)
	return hex.EncodeToString(bytes)
}

func generateToken() (string, error) {
	bytes := make([]byte, 32)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}

