package txmanager

import (
	"context"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type TxManager interface {
	WithTx(ctx context.Context, fn func(tx pgx.Tx) error) error
}

type PgxTxManager struct {
	db *pgxpool.Pool
}

func NewPgxTxManager(db *pgxpool.Pool) TxManager {
	return &PgxTxManager{
		db: db,
	}
}

func (m *PgxTxManager) WithTx(
	ctx context.Context,
	fn func(tx pgx.Tx) error,
) error {
	tx, err := m.db.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return err
	}

	defer func() {
		_ = tx.Rollback(ctx)
	}()

	if err := fn(tx); err != nil {
		return err
	}

	return tx.Commit(ctx)
}
