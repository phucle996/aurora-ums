package repoImple

import (
	"aurora/internal/domain/entity"
	domainrepo "aurora/internal/domain/repository"
	"aurora/internal/errorx"
	"aurora/internal/model"
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type MFARepoImple struct {
	db *pgxpool.Pool
}

func NewMFARepoImple(db *pgxpool.Pool) domainrepo.MFARepoInterface {
	return &MFARepoImple{db: db}
}

func (r *MFARepoImple) CreateMethod(ctx context.Context, method *entity.MFAMethod) error {
	if method == nil {
		return errorx.ErrEntityNil
	}
	m := model.MFAMethodEntityToModel(*method)

	const q = `
		INSERT INTO mfa_methods (
			id, user_id, method, secret, target, verified_at, created_at
		) VALUES ($1,$2,$3,$4,$5,$6,$7)
	`
	_, err := r.db.Exec(ctx, q,
		m.ID, m.UserID, m.Method, m.Secret, m.Target, m.VerifiedAt, m.CreatedAt,
	)
	return err
}

func (r *MFARepoImple) GetMethodByUserAndType(ctx context.Context, userID uuid.UUID, method entity.MFAMethodType) (*entity.MFAMethod, error) {
	const q = `
		SELECT id, user_id, method, secret, target, verified_at, created_at
		FROM mfa_methods
		WHERE user_id = $1 AND method = $2
		LIMIT 1
	`

	var m model.MFAMethod
	err := r.db.QueryRow(ctx, q, userID, method).Scan(
		&m.ID,
		&m.UserID,
		&m.Method,
		&m.Secret,
		&m.Target,
		&m.VerifiedAt,
		&m.CreatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, errorx.ErrMFAMethodNotFound
		}
		return nil, err
	}
	entityMethod := model.MFAMethodModelToEntity(m)
	return &entityMethod, nil
}

func (r *MFARepoImple) ListMethodsByUser(ctx context.Context, userID uuid.UUID) ([]entity.MFAMethod, error) {
	const q = `
		SELECT id, user_id, method, secret, target, verified_at, created_at
		FROM mfa_methods
		WHERE user_id = $1 AND verified_at IS NOT NULL
		ORDER BY created_at DESC
	`

	rows, err := r.db.Query(ctx, q, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := make([]entity.MFAMethod, 0)
	for rows.Next() {
		var m model.MFAMethod
		if err := rows.Scan(
			&m.ID,
			&m.UserID,
			&m.Method,
			&m.Secret,
			&m.Target,
			&m.VerifiedAt,
			&m.CreatedAt,
		); err != nil {
			return nil, err
		}
		out = append(out, model.MFAMethodModelToEntity(m))
	}
	if rows.Err() != nil {
		return nil, rows.Err()
	}
	return out, nil
}

func (r *MFARepoImple) UpdateMethodVerifiedAt(ctx context.Context, methodID uuid.UUID, verifiedAt time.Time) error {
	const q = `
		UPDATE mfa_methods
		SET verified_at = $1
		WHERE id = $2
	`
	cmd, err := r.db.Exec(ctx, q, verifiedAt, methodID)
	if err != nil {
		return err
	}
	if cmd.RowsAffected() == 0 {
		return errorx.ErrMFAMethodNotFound
	}
	return nil
}

func (r *MFARepoImple) DeleteMethod(ctx context.Context, userID uuid.UUID, method entity.MFAMethodType) error {
	const q = `
		DELETE FROM mfa_methods
		WHERE user_id = $1 AND method = $2
	`
	cmd, err := r.db.Exec(ctx, q, userID, method)
	if err != nil {
		return err
	}
	if cmd.RowsAffected() == 0 {
		return errorx.ErrMFAMethodNotFound
	}
	return nil
}

func (r *MFARepoImple) CreateRecoveryCodes(ctx context.Context, codes []entity.MFARecoveryCode) error {
	if len(codes) == 0 {
		return nil
	}

	batch := &pgx.Batch{}
	for _, code := range codes {
		m := model.MFARecoveryCodeEntityToModel(code)
		batch.Queue(
			`INSERT INTO mfa_recovery_codes (id, user_id, code_hash, created_at) VALUES ($1,$2,$3,$4)`,
			m.ID, m.UserID, m.CodeHash, m.CreatedAt,
		)
	}

	br := r.db.SendBatch(ctx, batch)
	defer br.Close()

	for i := 0; i < len(codes); i++ {
		if _, err := br.Exec(); err != nil {
			return err
		}
	}
	return nil
}

func (r *MFARepoImple) ListRecoveryCodesByUser(ctx context.Context, userID uuid.UUID) ([]entity.MFARecoveryCode, error) {
	const q = `
		SELECT id, user_id, code_hash, created_at
		FROM mfa_recovery_codes
		WHERE user_id = $1
		ORDER BY created_at DESC
	`

	rows, err := r.db.Query(ctx, q, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []entity.MFARecoveryCode
	for rows.Next() {
		var m model.MFARecoveryCode
		if err := rows.Scan(&m.ID, &m.UserID, &m.CodeHash, &m.CreatedAt); err != nil {
			return nil, err
		}
		out = append(out, model.MFARecoveryCodeToEntity(m))
	}
	if rows.Err() != nil {
		return nil, rows.Err()
	}
	return out, nil
}

func (r *MFARepoImple) DeleteRecoveryCodesByUser(ctx context.Context, userID uuid.UUID) error {
	const q = `
		DELETE FROM mfa_recovery_codes
		WHERE user_id = $1
	`
	_, err := r.db.Exec(ctx, q, userID)
	return err
}

func (r *MFARepoImple) CreateChallenge(ctx context.Context, challenge *entity.MFAChallenge) error {
	if challenge == nil {
		return errorx.ErrEntityNil
	}
	m := model.MFAChallengeEntityToModel(*challenge)

	const q = `
		INSERT INTO mfa_challenges (
			id, user_id, method, challenge_hash, expires_at, created_at
		) VALUES ($1,$2,$3,$4,$5,$6)
	`
	_, err := r.db.Exec(ctx, q,
		m.ID, m.UserID, m.Method, m.Challenge, m.ExpiresAt, m.CreatedAt,
	)
	return err
}

func (r *MFARepoImple) GetChallenge(ctx context.Context, challengeID uuid.UUID) (*entity.MFAChallenge, error) {
	const q = `
		SELECT id, user_id, method, challenge_hash, expires_at, created_at
		FROM mfa_challenges
		WHERE id = $1
		LIMIT 1
	`

	var m model.MFAChallenge
	err := r.db.QueryRow(ctx, q, challengeID).Scan(
		&m.ID,
		&m.UserID,
		&m.Method,
		&m.Challenge,
		&m.ExpiresAt,
		&m.CreatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, errorx.ErrMFAChallengeNotFound
		}
		return nil, err
	}
	entityChallenge := model.MFAChallengeModelToEntity(m)
	return &entityChallenge, nil
}

func (r *MFARepoImple) DeleteChallenge(ctx context.Context, challengeID uuid.UUID) error {
	const q = `
		DELETE FROM mfa_challenges
		WHERE id = $1
	`
	cmd, err := r.db.Exec(ctx, q, challengeID)
	if err != nil {
		return err
	}
	if cmd.RowsAffected() == 0 {
		return errorx.ErrMFAChallengeNotFound
	}
	return nil
}
