package domainrepo

import (
	"aurora/internal/domain/entity"
	"context"

	"github.com/google/uuid"
)

type RefreshRepoInterface interface {
	Create(ctx context.Context, refresh *entity.RefreshToken) error
	GetTokenByDevice(ctx context.Context, tokenHash string, deviceID uuid.UUID, deviceSecretHash string) (*entity.RefreshSession, error)
}
