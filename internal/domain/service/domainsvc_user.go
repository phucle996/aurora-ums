package domainsvc

import (
	"aurora/internal/domain/entity"
	"context"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

type UserSvcInterface interface {
	CreateUser(ctx context.Context, user *entity.User) error
	GetUserByUsername(ctx context.Context, username string) (*entity.User, error)
	GetUserByEmail(ctx context.Context, email string) (*entity.User, error)
	GetUserWithProfileByID(ctx context.Context, userID uuid.UUID) (*entity.UserWithProfile, error)
	GetCurrentUser(ctx context.Context, userID uuid.UUID) (*entity.CurrentUserView, error)
	UpsertProfile(ctx context.Context, profile *entity.Profile) error
	UpdatePasswordTx(ctx context.Context, tx pgx.Tx, userID uuid.UUID, passwordHash string) error
	UpdateUserStateTx(ctx context.Context, tx pgx.Tx, userID uuid.UUID, status entity.UserStatus) error
	GetUserStatusTx(ctx context.Context, tx pgx.Tx, userID uuid.UUID) (entity.UserStatus, error)
	ActiveAccount(ctx context.Context, userID uuid.UUID, token string) error
	VerifyResetPassword(ctx context.Context, userID uuid.UUID, token string) error
	NewPassword(ctx context.Context, userID uuid.UUID, token, password string) error
	ForgotPasswd(ctx context.Context, email string) error
}
