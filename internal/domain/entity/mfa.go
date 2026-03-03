package entity

import (
	"time"

	"github.com/google/uuid"
)

// MFAMethodType enumerates supported multi-factor mechanisms.
type MFAMethodType string

const (
	MFAMethodTOTP  MFAMethodType = "totp"
	MFAMethodSMS   MFAMethodType = "sms"
	MFAMethodEmail MFAMethodType = "email"
)

// MFAMethod represents a configured MFA factor for a user.
type MFAMethod struct {
	ID         uuid.UUID
	UserID     uuid.UUID
	Method     MFAMethodType
	Secret     *string
	Target     *string
	VerifiedAt *time.Time
	CreatedAt  *time.Time
}

// MFARecoveryCode holds recovery codes per user.
type MFARecoveryCode struct {
	ID        uuid.UUID
	UserID    uuid.UUID
	CodeHash  string
	CreatedAt *time.Time
}

// MFAChallenge tracks issued MFA challenges.
type MFAChallenge struct {
	ID        uuid.UUID
	UserID    uuid.UUID
	Method    MFAMethodType
	Challenge *string
	ExpiresAt *time.Time
	CreatedAt *time.Time
}
