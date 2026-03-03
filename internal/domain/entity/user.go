package entity

import (
	"time"

	"github.com/google/uuid"
)

// UserStatus represents the lifecycle state of a user.
type UserStatus string

const (
	UserStatusActive    UserStatus = "active"
	UserStatusPending   UserStatus = "pending"
	UserStatusSuspended UserStatus = "suspended"
	UserStatusDeleted   UserStatus = "deleted"
)

// User represents a platform account owner.
type User struct {
	ID         uuid.UUID
	Username   string
	Email      string
	Password   string
	Status     UserStatus
	UserLevel  int32
	OnBoarding bool
	CreatedAt  time.Time
	UpdatedAt  time.Time
}

// UserWithProfile aggregates user and profile data for read views.
type UserWithProfile struct {
	User    User
	Profile *Profile
}

// CurrentUserView is the scoped payload returned by /auth/me.
type CurrentUserView struct {
	User        User
	Profile     *Profile
	Roles       []string
	Permissions []string
}

// RefreshToken stores rotating refresh token metadata per device/user.
type RefreshToken struct {
	ID        uuid.UUID
	UserID    uuid.UUID
	DeviceID  uuid.UUID
	TokenHash string
	ExpiresAt time.Time
	CreatedAt time.Time
}

// UserDevice keeps track of a user device fingerprint.
type UserDevice struct {
	ID               uuid.UUID
	UserID           uuid.UUID
	DeviceID         uuid.UUID
	DeviceSecretHash string
	UserAgent        *string
	IPFirst          *string
	IPLast           *string
	Revoked          bool
	CreatedAt        *time.Time
	LastSeen         *time.Time
	RevokedAt        *time.Time
}
