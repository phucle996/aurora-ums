package entity

import (
	"time"

	"github.com/google/uuid"
)

// OneTimeTokenPurpose captures why a one-time token was issued.
type OttType string

const (
	OttAccountVerify OttType = "account_verify"
	OttPasswordReset OttType = "password_reset"
)

// OneTimeToken supports account verification and password reset flows.
type OneTimeToken struct {
	ID        uuid.UUID
	UserID    uuid.UUID
	TokenHash string
	Purpose   OttType
	ExpiresAt time.Time
	CreatedAt time.Time
}
