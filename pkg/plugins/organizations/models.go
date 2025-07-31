package organizations

import (
	"time"

	"gorm.io/gorm"
)

// Organization represents an organization with enhanced features
type Organization struct {
	ID          string         `gorm:"primaryKey"           json:"id"`
	Name        string         `gorm:"not null"             json:"name"`
	Slug        string         `gorm:"uniqueIndex;not null" json:"slug"`
	Logo        string         `                            json:"logo,omitempty"`
	Description string         `                            json:"description,omitempty"`
	Website     string         `                            json:"website,omitempty"`
	Industry    string         `                            json:"industry,omitempty"`
	Size        string         `                            json:"size,omitempty"`
	Timezone    string         `                            json:"timezone,omitempty"`
	Settings    map[string]any `gorm:"serializer:json"      json:"settings,omitempty"`
	Metadata    map[string]any `gorm:"serializer:json"      json:"metadata,omitempty"`
	CreatedAt   time.Time      `                            json:"createdAt"`
	UpdatedAt   time.Time      `                            json:"updatedAt"`
	DeletedAt   gorm.DeletedAt `gorm:"index"                json:"-"`

	// Relationships
	Members     []OrganizationMember     `gorm:"foreignKey:OrganizationID" json:"members,omitempty"`
	Teams       []Team                   `gorm:"foreignKey:OrganizationID" json:"teams,omitempty"`
	Invitations []OrganizationInvitation `gorm:"foreignKey:OrganizationID" json:"invitations,omitempty"`
}

// OrganizationMember represents a user's membership in an organization
type OrganizationMember struct {
	ID             string         `gorm:"primaryKey"                json:"id"`
	OrganizationID string         `gorm:"not null;index"            json:"organizationId"`
	UserID         string         `gorm:"not null;index"            json:"userId"`
	Role           string         `gorm:"not null;default:'member'" json:"role"`
	Status         string         `gorm:"not null;default:'active'" json:"status"`
	Permissions    []string       `gorm:"serializer:json"           json:"permissions,omitempty"`
	JoinedAt       time.Time      `                                 json:"joinedAt"`
	CreatedAt      time.Time      `                                 json:"createdAt"`
	UpdatedAt      time.Time      `                                 json:"updatedAt"`
	DeletedAt      gorm.DeletedAt `gorm:"index"                     json:"-"`

	// Relationships
	Organization *Organization `gorm:"foreignKey:OrganizationID" json:"organization,omitempty"`
}

// Team represents a team within an organization
type Team struct {
	ID             string         `gorm:"primaryKey"      json:"id"`
	OrganizationID string         `gorm:"not null;index"  json:"organizationId"`
	Name           string         `gorm:"not null"        json:"name"`
	Slug           string         `gorm:"not null"        json:"slug"`
	Description    string         `                       json:"description,omitempty"`
	Color          string         `                       json:"color,omitempty"`
	IsDefault      bool           `gorm:"default:false"   json:"isDefault"`
	Settings       map[string]any `gorm:"serializer:json" json:"settings,omitempty"`
	Metadata       map[string]any `gorm:"serializer:json" json:"metadata,omitempty"`
	CreatedAt      time.Time      `                       json:"createdAt"`
	UpdatedAt      time.Time      `                       json:"updatedAt"`
	DeletedAt      gorm.DeletedAt `gorm:"index"           json:"-"`

	// Relationships
	Organization *Organization `gorm:"foreignKey:OrganizationID" json:"organization,omitempty"`
	Members      []TeamMember  `gorm:"foreignKey:TeamID"         json:"members,omitempty"`
}

// TeamMember represents a user's membership in a team
type TeamMember struct {
	ID        string         `gorm:"primaryKey"                json:"id"`
	TeamID    string         `gorm:"not null;index"            json:"teamId"`
	UserID    string         `gorm:"not null;index"            json:"userId"`
	Role      string         `gorm:"not null;default:'member'" json:"role"`
	JoinedAt  time.Time      `                                 json:"joinedAt"`
	CreatedAt time.Time      `                                 json:"createdAt"`
	UpdatedAt time.Time      `                                 json:"updatedAt"`
	DeletedAt gorm.DeletedAt `gorm:"index"                     json:"-"`

	// Relationships
	Team *Team `gorm:"foreignKey:TeamID" json:"team,omitempty"`
}

// OrganizationInvitation represents an invitation to join an organization
type OrganizationInvitation struct {
	ID             string         `gorm:"primaryKey"                 json:"id"`
	OrganizationID string         `gorm:"not null;index"             json:"organizationId"`
	Email          string         `gorm:"not null;index"             json:"email"`
	Role           string         `gorm:"not null;default:'member'"  json:"role"`
	Status         string         `gorm:"not null;default:'pending'" json:"status"`
	Token          string         `gorm:"uniqueIndex;not null"       json:"-"`
	InvitedBy      string         `gorm:"not null"                   json:"invitedBy"`
	Message        string         `                                  json:"message,omitempty"`
	ExpiresAt      time.Time      `                                  json:"expiresAt"`
	AcceptedAt     *time.Time     `                                  json:"acceptedAt,omitempty"`
	RejectedAt     *time.Time     `                                  json:"rejectedAt,omitempty"`
	Metadata       map[string]any `gorm:"serializer:json"            json:"metadata,omitempty"`
	CreatedAt      time.Time      `                                  json:"createdAt"`
	UpdatedAt      time.Time      `                                  json:"updatedAt"`
	DeletedAt      gorm.DeletedAt `gorm:"index"                      json:"-"`

	// Relationships
	Organization *Organization `gorm:"foreignKey:OrganizationID" json:"organization,omitempty"`
}

// OrganizationRole defines the available roles in an organization
type OrganizationRole string

const (
	RoleOwner  OrganizationRole = "owner"
	RoleAdmin  OrganizationRole = "admin"
	RoleMember OrganizationRole = "member"
	RoleGuest  OrganizationRole = "guest"
)

// TeamRole defines the available roles in a team
type TeamRole string

const (
	TeamRoleManager TeamRole = "manager"
	TeamRoleMember  TeamRole = "member"
)

// InvitationStatus defines the possible statuses of an invitation
type InvitationStatus string

const (
	InvitationPending  InvitationStatus = "pending"
	InvitationAccepted InvitationStatus = "accepted"
	InvitationRejected InvitationStatus = "rejected"
	InvitationExpired  InvitationStatus = "expired"
)

// MemberStatus defines the possible statuses of a member
type MemberStatus string

const (
	MemberActive    MemberStatus = "active"
	MemberInactive  MemberStatus = "inactive"
	MemberSuspended MemberStatus = "suspended"
)

// Permission constants
const (
	PermissionReadOrganization   = "org:read"
	PermissionWriteOrganization  = "org:write"
	PermissionDeleteOrganization = "org:delete"
	PermissionManageMembers      = "members:manage"
	PermissionInviteMembers      = "members:invite"
	PermissionRemoveMembers      = "members:remove"
	PermissionManageTeams        = "teams:manage"
	PermissionCreateTeams        = "teams:create"
	PermissionDeleteTeams        = "teams:delete"
	PermissionManageSettings     = "settings:manage"
)

// Default permissions for each role
var DefaultPermissions = map[OrganizationRole][]string{
	RoleOwner: {
		PermissionReadOrganization,
		PermissionWriteOrganization,
		PermissionDeleteOrganization,
		PermissionManageMembers,
		PermissionInviteMembers,
		PermissionRemoveMembers,
		PermissionManageTeams,
		PermissionCreateTeams,
		PermissionDeleteTeams,
		PermissionManageSettings,
	},
	RoleAdmin: {
		PermissionReadOrganization,
		PermissionWriteOrganization,
		PermissionManageMembers,
		PermissionInviteMembers,
		PermissionRemoveMembers,
		PermissionManageTeams,
		PermissionCreateTeams,
		PermissionDeleteTeams,
		PermissionManageSettings,
	},
	RoleMember: {
		PermissionReadOrganization,
	},
	RoleGuest: {
		PermissionReadOrganization,
	},
}

// TableName overrides the table name for Organization
func (Organization) TableName() string {
	return "organizations"
}

// TableName overrides the table name for OrganizationMember
func (OrganizationMember) TableName() string {
	return "organization_members"
}

// TableName overrides the table name for Team
func (Team) TableName() string {
	return "teams"
}

// TableName overrides the table name for TeamMember
func (TeamMember) TableName() string {
	return "team_members"
}

// TableName overrides the table name for OrganizationInvitation
func (OrganizationInvitation) TableName() string {
	return "organization_invitations"
}

