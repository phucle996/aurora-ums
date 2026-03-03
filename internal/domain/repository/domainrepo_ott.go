package domainrepo

import (
	"aurora/internal/domain/entity"
	"context"

	"github.com/jackc/pgx/v5"
)

type OttRepoInterface interface {
	Create(ctx context.Context, ott *entity.OneTimeToken) error
	Validate(ctx context.Context, ott *entity.OneTimeToken) error
	ConsumTx(ctx context.Context, tx pgx.Tx, ott *entity.OneTimeToken) error
}
