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
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

type UserRepoImple struct {
	db *pgxpool.Pool
}

func NewUserRepoImple(db *pgxpool.Pool) domainrepo.UserRepoInterface {
	return &UserRepoImple{
		db: db,
	}

}

func (r *UserRepoImple) CheckUserExist(ctx context.Context, username, email string) error {

	const q = `
		SELECT 1
		FROM users
		WHERE (username = $1 OR email = $2)
		LIMIT 1
	`

	var exists int
	err := r.db.QueryRow(ctx, q, username, email).Scan(&exists)

	if err == nil {
		// user tồn tại
		return errorx.ErrUserAlreadyExist
	}

	if errors.Is(err, pgx.ErrNoRows) {
		// user chưa tồn tại
		return nil
	}

	// lỗi hệ thống
	return err
}

func (r *UserRepoImple) CreateUser(ctx context.Context, user *entity.User) error {
	mUser := model.UserEntityToModel(*user)

	tx, err := r.db.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	const insertUser = `
		INSERT INTO users (
		  id, username, email, password_hash, status, user_level, on_boarding, created_at, updated_at
		)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9)
	`

	_, err = tx.Exec(ctx, insertUser,
		mUser.ID,
		mUser.Username,
		mUser.Email,
		mUser.Password,
		mUser.Status,
		mUser.UserLevel,
		mUser.OnBoarding,
		mUser.CreatedAt,
		mUser.UpdatedAt,
	)
	if err != nil {
		return err
	}

	roleName := "user"
	switch {
	case mUser.UserLevel <= 1:
		roleName = "root"
	case mUser.UserLevel <= 3:
		roleName = "admin"
	}

	const insertUserRole = `
		INSERT INTO user_roles (id, user_id, role_id, created_at)
		SELECT $1, $2, r.id, $3
		FROM roles r
		WHERE r.name = $4
		  AND r.scope = 'global'
		LIMIT 1
	`

	assignedAt := time.Now().UTC()
	cmd, err := tx.Exec(ctx, insertUserRole,
		uuid.New(),
		mUser.ID,
		assignedAt,
		roleName,
	)
	if err != nil {
		return err
	}
	if cmd.RowsAffected() == 0 {
		return errorx.ErrRoleNotFound
	}

	if err := tx.Commit(ctx); err != nil {
		return err
	}

	return nil
}

func (r *UserRepoImple) GetUserByUsername(ctx context.Context, username string) (*entity.User, error) {
	const q = `
		SELECT
			u.id,
			u.username,
			u.email,
			u.password_hash,
			u.status,
			u.user_level,
			u.on_boarding,
			u.created_at,
			u.updated_at
		FROM users u
		WHERE u.username = $1
		LIMIT 1
	`

	var m model.User
	err := r.db.QueryRow(ctx, q, username).Scan(
		&m.ID,
		&m.Username,
		&m.Email,
		&m.Password,
		&m.Status,
		&m.UserLevel,
		&m.OnBoarding,
		&m.CreatedAt,
		&m.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, errorx.ErrUserNotFound
		}
		return nil, err
	}

	user := model.UserModelToEntity(m)
	return &user, nil
}

func (r *UserRepoImple) GetUserByEmail(ctx context.Context, email string) (*entity.User, error) {
	const q = `
		SELECT
			u.id,
			u.username,
			u.email,
			u.password_hash,
			u.status,
			u.user_level,
			u.on_boarding,
			u.created_at,
			u.updated_at
		FROM users u
		WHERE u.email = $1
		LIMIT 1
	`

	var m model.User
	err := r.db.QueryRow(ctx, q, email).Scan(
		&m.ID,
		&m.Username,
		&m.Email,
		&m.Password,
		&m.Status,
		&m.UserLevel,
		&m.OnBoarding,
		&m.CreatedAt,
		&m.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, errorx.ErrUserNotFound
		}
		return nil, err
	}

	user := model.UserModelToEntity(m)
	return &user, nil
}

func (r *UserRepoImple) GetUserByID(ctx context.Context, userID uuid.UUID) (*entity.User, error) {
	const q = `
		SELECT
			u.id,
			u.username,
			u.email,
			u.password_hash,
			u.status,
			u.user_level,
			u.on_boarding,
			u.created_at,
			u.updated_at
		FROM users u
		WHERE u.id = $1
		LIMIT 1
	`

	var m model.User
	err := r.db.QueryRow(ctx, q, userID).Scan(
		&m.ID,
		&m.Username,
		&m.Email,
		&m.Password,
		&m.Status,
		&m.UserLevel,
		&m.OnBoarding,
		&m.CreatedAt,
		&m.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, errorx.ErrUserNotFound
		}
		return nil, err
	}

	user := model.UserModelToEntity(m)
	return &user, nil
}

func (r *UserRepoImple) GetProfileByUserID(ctx context.Context, userID uuid.UUID) (*entity.Profile, error) {
	const q = `
		SELECT id, user_id, full_name, company, referral_source, phone, job_function, country, avatar_url, bio, created_at, updated_at
		FROM profiles
		WHERE user_id = $1
		LIMIT 1
	`

	var m model.Profile
	err := r.db.QueryRow(ctx, q, userID).Scan(
		&m.ID,
		&m.UserID,
		&m.FullName,
		&m.Company,
		&m.ReferralSource,
		&m.Phone,
		&m.JobFunction,
		&m.Country,
		&m.AvatarURL,
		&m.Bio,
		&m.CreatedAt,
		&m.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, errorx.ErrProfileNotFound
		}
		return nil, err
	}

	profile := model.ProfileModelToEntity(m)
	return &profile, nil
}

func (r *UserRepoImple) GetUserWithProfileByID(
	ctx context.Context,
	userID uuid.UUID,
) (*entity.UserWithProfile, error) {

	const q = `
	SELECT
		u.id,
		u.username,
		u.email,
		u.password_hash,
		u.status,
		u.user_level,
		u.on_boarding,

		p.id,
		p.user_id,
		p.full_name,
		p.company,
		p.referral_source,
		p.phone,
		p.job_function,
		p.country,
		p.avatar_url,
		p.bio

	FROM users u
	LEFT JOIN profiles p ON p.user_id = u.id
	WHERE u.id = $1
	LIMIT 1
	`

	var mUser model.User
	var mProfile model.Profile

	err := r.db.QueryRow(ctx, q, userID).Scan(
		&mUser.ID,
		&mUser.Username,
		&mUser.Email,
		&mUser.Password,
		&mUser.Status,
		&mUser.UserLevel,
		&mUser.OnBoarding,

		&mProfile.ID,
		&mProfile.UserID,
		&mProfile.FullName,
		&mProfile.Company,
		&mProfile.ReferralSource,
		&mProfile.Phone,
		&mProfile.JobFunction,
		&mProfile.Country,
		&mProfile.AvatarURL,
		&mProfile.Bio,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, errorx.ErrUserNotFound
		}
		return nil, err
	}

	user := model.UserModelToEntity(mUser)

	profile := model.ProfileModelToEntity(mProfile)

	return &entity.UserWithProfile{
		User:    user,
		Profile: &profile,
	}, nil
}

func (r *UserRepoImple) UpsertProfile(ctx context.Context, profile *entity.Profile) error {
	if profile == nil {
		return errorx.ErrEntityNil
	}
	m := model.ProfileEntityToModel(*profile)

	const q = `
		WITH upserted AS (
			INSERT INTO profiles (
				id,
				user_id,
				full_name,
				company,
				referral_source,
				phone,
				job_function,
				country,
				avatar_url,
				bio,
				created_at,
				updated_at
			) VALUES (
				$1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12
			)
			ON CONFLICT (user_id) DO UPDATE SET
				full_name = EXCLUDED.full_name,
				company = EXCLUDED.company,
				referral_source = EXCLUDED.referral_source,
				phone = EXCLUDED.phone,
				job_function = EXCLUDED.job_function,
				country = EXCLUDED.country,
				avatar_url = EXCLUDED.avatar_url,
				bio = EXCLUDED.bio,
				updated_at = EXCLUDED.updated_at
			RETURNING user_id, avatar_url
		)
		UPDATE users
		SET on_boarding = TRUE,
		    updated_at = $12
		FROM upserted
		WHERE users.id = upserted.user_id
	`

	_, err := r.db.Exec(ctx, q,
		m.ID,
		m.UserID,
		m.FullName,
		m.Company,
		m.ReferralSource,
		m.Phone,
		m.JobFunction,
		m.Country,
		m.AvatarURL,
		m.Bio,
		m.CreatedAt,
		m.UpdatedAt,
	)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23503" {
			return errorx.ErrUserNotFound
		}
		return err
	}
	return nil
}

func (r *UserRepoImple) UpdateStatusUserTx(ctx context.Context, tx pgx.Tx, userID uuid.UUID, status entity.UserStatus) error {
	const q = `
		UPDATE users
		SET status = $1, updated_at = $2
		WHERE id = $3
	`

	now := time.Now().UTC()
	cmd, err := tx.Exec(ctx, q, status, now, userID)
	if err != nil {
		return err
	}

	affected := cmd.RowsAffected()
	if affected == 0 {
		return errorx.ErrUserNotFound
	}
	if affected > 1 {
		return errorx.ErrUnexpectedRows
	}
	return nil
}

func (r *UserRepoImple) UpdatePasswordTx(ctx context.Context, tx pgx.Tx, userID uuid.UUID, passwordHash string) error {
	const q = `
		UPDATE users
		SET password_hash = $1, updated_at = $2
		WHERE id = $3
	`

	now := time.Now().UTC()
	cmd, err := tx.Exec(ctx, q, passwordHash, now, userID)
	if err != nil {
		return err
	}

	affected := cmd.RowsAffected()
	if affected == 0 {
		return errorx.ErrUserNotFound
	}
	if affected > 1 {
		return errorx.ErrUnexpectedRows
	}
	return nil
}

func (r *UserRepoImple) GetUserStatusTx(ctx context.Context, tx pgx.Tx, userID uuid.UUID) (entity.UserStatus, error) {
	const q = `
		SELECT status
		FROM users
		WHERE id = $1
		LIMIT 1
	`

	var status entity.UserStatus
	if err := tx.QueryRow(ctx, q, userID).Scan(&status); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return "", errorx.ErrUserNotFound
		}
		return "", err
	}
	return status, nil
}
