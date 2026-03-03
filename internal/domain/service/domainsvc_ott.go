package domainsvc

import (
	"aurora/internal/domain/entity"
	"context"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

type OttSvcInterface interface {
	CreateToken(ctx context.Context, userID uuid.UUID, tokenType entity.OttType) (string, error)
	ValidateToken(ctx context.Context, userID uuid.UUID, plainToken string, tokenType entity.OttType) error
	ConsumTokenTx(ctx context.Context, tx pgx.Tx, userID uuid.UUID, plainToken string, tokenType entity.OttType) error
}
