package domainrepo

import (
	"aurora/internal/domain/entity"
	"context"
	"time"

	"github.com/google/uuid"
)

type MFARepoInterface interface {
	CreateMethod(ctx context.Context, method *entity.MFAMethod) error
	GetMethodByUserAndType(ctx context.Context, userID uuid.UUID, method entity.MFAMethodType) (*entity.MFAMethod, error)
	ListMethodsByUser(ctx context.Context, userID uuid.UUID) ([]entity.MFAMethod, error)
	UpdateMethodVerifiedAt(ctx context.Context, methodID uuid.UUID, verifiedAt time.Time) error
	DeleteMethod(ctx context.Context, userID uuid.UUID, method entity.MFAMethodType) error

	CreateRecoveryCodes(ctx context.Context, codes []entity.MFARecoveryCode) error
	ListRecoveryCodesByUser(ctx context.Context, userID uuid.UUID) ([]entity.MFARecoveryCode, error)
	DeleteRecoveryCodesByUser(ctx context.Context, userID uuid.UUID) error

	CreateChallenge(ctx context.Context, challenge *entity.MFAChallenge) error
	GetChallenge(ctx context.Context, challengeID uuid.UUID) (*entity.MFAChallenge, error)
	DeleteChallenge(ctx context.Context, challengeID uuid.UUID) error
}
