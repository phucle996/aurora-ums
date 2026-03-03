package repoImple

import (
	"aurora/internal/domain/entity"
	domainrepo "aurora/internal/domain/repository"
	"aurora/internal/errorx"
	"aurora/internal/model"
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type OttRepoImple struct {
	db *pgxpool.Pool
}

func NewOttRepoImple(db *pgxpool.Pool) domainrepo.OttRepoInterface {
	return &OttRepoImple{
		db: db,
	}

}

func (r *OttRepoImple) Create(ctx context.Context, ott *entity.OneTimeToken) error {
	if ott == nil {
		return errorx.ErrEntityNil
	}

	m := model.OttEntityToModel(*ott)

	const q = `
	INSERT INTO one_time_tokens (
		id,
		user_id,
		token_hash,
		purpose,
		expires_at,
		created_at
	)
	VALUES ($1, $2, $3, $4, $5, $6)
	ON CONFLICT (user_id, purpose)
	DO UPDATE SET
		id = EXCLUDED.id,
		token_hash = EXCLUDED.token_hash,
		expires_at = EXCLUDED.expires_at,
		created_at = EXCLUDED.created_at
`

	_, err := r.db.Exec(ctx, q,
		m.ID,
		m.UserID,
		m.TokenHash,
		m.Purpose,
		m.ExpiresAt,
		m.CreatedAt,
	)
	if err != nil {
		return err
	}

	ott.ID = m.ID
	ott.CreatedAt = m.CreatedAt
	return nil
}

func (r *OttRepoImple) Consum(ctx context.Context, ott *entity.OneTimeToken) error {
	m := model.OttEntityToModel(*ott)

	const q = `
		DELETE FROM one_time_tokens
		WHERE token_hash = $1 AND user_id = $2 AND purpose = $3
		  AND (expires_at IS NULL OR expires_at > NOW())
	`

	cmd, err := r.db.Exec(ctx, q, m.TokenHash, m.UserID, m.Purpose)
	if err != nil {
		return err
	}

	affected := cmd.RowsAffected()
	if affected == 0 {
		return errorx.ErrOttNotFound
	}
	if affected > 1 {
		return errorx.ErrUnexpectedRows
	}
	return nil
}

func (r *OttRepoImple) Validate(ctx context.Context, ott *entity.OneTimeToken) error {
	m := model.OttEntityToModel(*ott)

	const q = `
		SELECT 1
		FROM one_time_tokens
		WHERE token_hash = $1 AND user_id = $2 AND purpose = $3
		  AND expires_at > NOW()
		LIMIT 1
	`

	var exists int
	err := r.db.QueryRow(ctx, q, m.TokenHash, m.UserID, m.Purpose).Scan(&exists)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return errorx.ErrOttNotFound
		}
		return err
	}
	return nil
}

func (r *OttRepoImple) ConsumTx(ctx context.Context, tx pgx.Tx, ott *entity.OneTimeToken) error {
	m := model.OttEntityToModel(*ott)

	const q = `
		DELETE FROM one_time_tokens
		WHERE token_hash = $1 AND user_id = $2 AND purpose = $3
		  AND (expires_at IS NULL OR expires_at > NOW())
	`

	cmd, err := tx.Exec(ctx, q, m.TokenHash, m.UserID, m.Purpose)
	if err != nil {
		return err
	}

	affected := cmd.RowsAffected()
	if affected == 0 {
		return errorx.ErrOttNotFound
	}
	if affected > 1 {
		return errorx.ErrUnexpectedRows
	}
	return nil
}
