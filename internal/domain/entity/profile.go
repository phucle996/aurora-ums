package entity

import (
	"time"

	"github.com/google/uuid"
)

// Profile stores user profile information separate from credentials.
type Profile struct {
	ID             uuid.UUID
	UserID         uuid.UUID
	FullName       *string
	Company        *string
	ReferralSource *string
	Phone          *string
	JobFunction    *string
	Country        *string
	AvatarURL      *string
	Bio            *string
	CreatedAt      *time.Time
	UpdatedAt      *time.Time
}
