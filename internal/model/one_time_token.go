package model

import (
	"time"

	"aurora/internal/domain/entity"

	"github.com/google/uuid"
)

// OneTimeToken mirrors one_time_tokens table.
type OneTimeToken struct {
	ID        uuid.UUID      `db:"id"`
	UserID    uuid.UUID      `db:"user_id"`
	TokenHash string         `db:"token_hash"`
	Purpose   entity.OttType `db:"purpose"`
	ExpiresAt time.Time      `db:"expires_at"`
	CreatedAt time.Time      `db:"created_at"`
}

func OttModelToEntity(m OneTimeToken) entity.OneTimeToken {
	return entity.OneTimeToken{
		ID:        m.ID,
		UserID:    m.UserID,
		TokenHash: m.TokenHash,
		Purpose:   entity.OttType(m.Purpose),
		ExpiresAt: m.ExpiresAt,
		CreatedAt: m.CreatedAt,
	}
}

func OttEntityToModel(e entity.OneTimeToken) OneTimeToken {
	return OneTimeToken{
		ID:        e.ID,
		UserID:    e.UserID,
		TokenHash: e.TokenHash,
		Purpose:   entity.OttType(e.Purpose),
		ExpiresAt: e.ExpiresAt,
		CreatedAt: e.CreatedAt,
	}
}
