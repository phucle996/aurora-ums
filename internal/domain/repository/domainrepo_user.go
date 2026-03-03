package domainrepo

import (
	"aurora/internal/domain/entity"
	"context"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

type UserRepoInterface interface {
	CheckUserExist(ctx context.Context, username, email string) error
	CreateUser(ctx context.Context, user *entity.User) error
	GetUserByUsername(ctx context.Context, username string) (*entity.User, error)
	GetUserByEmail(ctx context.Context, email string) (*entity.User, error)
	GetUserByID(ctx context.Context, userID uuid.UUID) (*entity.User, error)
	GetUserWithProfileByID(ctx context.Context, userID uuid.UUID) (*entity.UserWithProfile, error)
	UpsertProfile(ctx context.Context, profile *entity.Profile) error
	UpdatePasswordTx(ctx context.Context, tx pgx.Tx, userID uuid.UUID, passwordHash string) error
	UpdateStatusUserTx(ctx context.Context, tx pgx.Tx, userID uuid.UUID, status entity.UserStatus) error
	GetUserStatusTx(ctx context.Context, tx pgx.Tx, userID uuid.UUID) (entity.UserStatus, error)
}
