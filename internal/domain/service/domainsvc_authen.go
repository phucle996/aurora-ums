package domainsvc

import (
	"aurora/internal/domain/entity"
	"context"

	"github.com/google/uuid"
)

type AuthSvcInterface interface {
	RegisterAccount(ctx context.Context, user *entity.User) error
	Login(ctx context.Context, username, passwd string) (*entity.LoginWithPasswd, error)
	VerifyMFAChallenge(ctx context.Context, userID uuid.UUID, sessionToken string, method entity.MFAMethodType, code string) (*entity.LoginWithPasswd, error)
	Refresh(ctx context.Context, refresh string) (*entity.LoginWithPasswd, error)
	Logout(ctx context.Context) error
}
