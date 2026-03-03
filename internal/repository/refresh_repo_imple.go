package repoImple

import (
	"aurora/internal/domain/entity"
	domainrepo "aurora/internal/domain/repository"
	"aurora/internal/errorx"
	"aurora/internal/model"
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type RefreshRepoImple struct {
	db *pgxpool.Pool
}

func NewRefreshRepoImple(db *pgxpool.Pool) domainrepo.RefreshRepoInterface {
	return &RefreshRepoImple{
		db: db,
	}
}

func (r *RefreshRepoImple) Create(ctx context.Context, refresh *entity.RefreshToken) error {
	if refresh == nil {
		return errorx.ErrEntityNil
	}

	m := model.RefreshTokenEntityToModel(*refresh)

	const q = `
		INSERT INTO refresh_tokens (
			id,
			user_id,
			device_id,
			token_hash,
			expires_at,
			created_at
		)
		VALUES ($1, $2, $3, $4, $5, $6)
	`

	_, err := r.db.Exec(ctx, q,
		m.ID,
		m.UserID,
		m.DeviceID,
		m.TokenHash,
		m.ExpiresAt,
		m.CreatedAt,
	)
	if err != nil {
		return err
	}

	return nil
}

func (r *RefreshRepoImple) GetTokenByDevice(ctx context.Context, tokenHash string, deviceID uuid.UUID, deviceSecretHash string) (*entity.RefreshSession, error) {
	const q = `
		SELECT
			rt.id,
			rt.user_id,
			rt.device_id,
			rt.token_hash,
			rt.expires_at,
			rt.created_at,
			u.status,
			u.username
		FROM refresh_tokens rt
		JOIN users u ON u.id = rt.user_id
		JOIN user_devices ud ON ud.device_id = rt.device_id
		WHERE rt.token_hash = $1
		  AND rt.device_id = $2
		  AND ud.device_secret_hash = $3
		  AND ud.revoked = FALSE
		  AND rt.expires_at > NOW()
		LIMIT 1
	`

	var m model.RefreshToken
	var status entity.UserStatus
	var username string
	err := r.db.QueryRow(ctx, q, tokenHash, deviceID, deviceSecretHash).Scan(
		&m.ID,
		&m.UserID,
		&m.DeviceID,
		&m.TokenHash,
		&m.ExpiresAt,
		&m.CreatedAt,
		&status,
		&username,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, errorx.ErrTokenInvalid
		}
		return nil, err
	}

	return &entity.RefreshSession{
		RefreshToken: model.RefreshTokenModelToEntity(m),
		UserID:       m.UserID,
		DeviceID:     m.DeviceID,
		Status:       status,
		Username:     username,
	}, nil
}
