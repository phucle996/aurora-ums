package domainsvc

import (
	"aurora/internal/domain/entity"
	"context"

	"github.com/google/uuid"
)

type RefreshSvcInterface interface {
	Create(ctx context.Context, refresh *entity.RefreshToken, plainToken string) error
	GetRefreshTokenByDevice(ctx context.Context, refresh string, deviceID uuid.UUID,
		deviceSecretHash string) (*entity.RefreshSession, error)
}
