package model

import (
	"time"

	"aurora/internal/domain/entity"

	"github.com/google/uuid"
)

// MFAMethod mirrors mfa_methods table.
type MFAMethod struct {
	ID         uuid.UUID            `db:"id"`
	UserID     uuid.UUID            `db:"user_id"`
	Method     entity.MFAMethodType `db:"method"`
	Secret     *string              `db:"secret"`
	Target     *string              `db:"target"`
	VerifiedAt *time.Time           `db:"verified_at"`
	CreatedAt  *time.Time           `db:"created_at"`
}

// MFARecoveryCode mirrors mfa_recovery_codes table.
type MFARecoveryCode struct {
	ID        uuid.UUID  `db:"id"`
	UserID    uuid.UUID  `db:"user_id"`
	CodeHash  string     `db:"code_hash"`
	CreatedAt *time.Time `db:"created_at"`
}

// MFAChallenge mirrors mfa_challenges table.
type MFAChallenge struct {
	ID        uuid.UUID            `db:"id"`
	UserID    uuid.UUID            `db:"user_id"`
	Method    entity.MFAMethodType `db:"method"`
	Challenge *string              `db:"challenge_hash"`
	ExpiresAt *time.Time           `db:"expires_at"`
	CreatedAt *time.Time           `db:"created_at"`
}

func MFAMethodModelToEntity(m MFAMethod) entity.MFAMethod {
	return entity.MFAMethod{
		ID:         m.ID,
		UserID:     m.UserID,
		Method:     m.Method,
		Secret:     m.Secret,
		Target:     m.Target,
		VerifiedAt: m.VerifiedAt,
		CreatedAt:  m.CreatedAt,
	}
}

func MFAMethodEntityToModel(e entity.MFAMethod) MFAMethod {
	return MFAMethod{
		ID:         e.ID,
		UserID:     e.UserID,
		Method:     e.Method,
		Secret:     e.Secret,
		Target:     e.Target,
		VerifiedAt: e.VerifiedAt,
		CreatedAt:  e.CreatedAt,
	}
}

func MFARecoveryCodeToEntity(m MFARecoveryCode) entity.MFARecoveryCode {
	return entity.MFARecoveryCode{
		ID:        m.ID,
		UserID:    m.UserID,
		CodeHash:  m.CodeHash,
		CreatedAt: m.CreatedAt,
	}
}

func MFARecoveryCodeEntityToModel(e entity.MFARecoveryCode) MFARecoveryCode {
	return MFARecoveryCode{
		ID:        e.ID,
		UserID:    e.UserID,
		CodeHash:  e.CodeHash,
		CreatedAt: e.CreatedAt,
	}
}

func MFAChallengeModelToEntity(m MFAChallenge) entity.MFAChallenge {
	return entity.MFAChallenge{
		ID:        m.ID,
		UserID:    m.UserID,
		Method:    m.Method,
		Challenge: m.Challenge,
		ExpiresAt: m.ExpiresAt,
		CreatedAt: m.CreatedAt,
	}
}

func MFAChallengeEntityToModel(e entity.MFAChallenge) MFAChallenge {
	return MFAChallenge{
		ID:        e.ID,
		UserID:    e.UserID,
		Method:    e.Method,
		Challenge: e.Challenge,
		ExpiresAt: e.ExpiresAt,
		CreatedAt: e.CreatedAt,
	}
}
