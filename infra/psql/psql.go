package psqlinfra

import (
	"aurora/internal/config"
	"aurora/pkg/logger"
	"context"
	"fmt"
	"net/url"
	"strings"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

func NewPostgres(
	ctx context.Context,
	psqlCfg *config.PsqlCfg,
) (*pgxpool.Pool, error) {

	const op = "system.postgres.connect"

	const (
		maxRetries    = 10
		retryInterval = 3 * time.Second
		pingTimeout   = 3 * time.Second
	)

	dsn, err := buildPostgresDSN(psqlCfg)
	if err != nil {
		logger.SysError(op, err, "invalid postgres configuration")
		return nil, err
	}

	var lastErr error

	for attempt := 1; attempt <= maxRetries; attempt++ {

		if ctx.Err() != nil {
			return nil, ctx.Err()
		}

		poolCfg, err := pgxpool.ParseConfig(dsn)
		if err != nil {
			// config lỗi → fail fast
			return nil, fmt.Errorf("parse pgx config: %w", err)
		}
		schema := strings.TrimSpace(psqlCfg.Schema)
		if schema == "" {
			schema = "ums"
		}
		poolCfg.ConnConfig.RuntimeParams["search_path"] = schema

		pool, err := pgxpool.NewWithConfig(ctx, poolCfg)
		if err != nil {
			lastErr = err
			logger.SysError(
				op, err,
				"attempt %d/%d: failed to create pool",
				attempt, maxRetries,
			)
			time.Sleep(retryInterval)
			continue
		}

		pingCtx, cancel := context.WithTimeout(ctx, pingTimeout)
		err = pool.Ping(pingCtx)
		cancel()

		if err == nil {
			logger.SysInfo(
				op,
				"connected to PostgreSQL (TLS %s)",
				psqlCfg.SslMode,
			)
			return pool, nil
		}

		lastErr = err
		logger.SysError(
			op, err,
			"attempt %d/%d: failed to ping DB",
			attempt, maxRetries,
		)

		pool.Close()

		if attempt < maxRetries {
			time.Sleep(retryInterval)
		}
	}

	return nil, fmt.Errorf(
		"could not connect to PostgreSQL after %d attempts: %w",
		maxRetries, lastErr,
	)
}

func buildPostgresDSN(cfg *config.PsqlCfg) (string, error) {
	u, err := url.Parse(cfg.DbURL)
	if err != nil {
		return "", fmt.Errorf("invalid DATABASE_URL: %w", err)
	}

	// Validate scheme
	if u.Scheme != "postgres" && u.Scheme != "postgresql" {
		return "", fmt.Errorf("unsupported DB scheme: %s", u.Scheme)
	}

	q := u.Query()

	if cfg.SslMode != "" {
		q.Set("sslmode", cfg.SslMode)
	}

	if cfg.CA != "" {
		q.Set("sslrootcert", cfg.CA)
	}

	// mTLS
	if cfg.ClientCert != "" {
		q.Set("sslcert", cfg.ClientCert)
	}
	if cfg.ClientKey != "" {
		q.Set("sslkey", cfg.ClientKey)
	}

	u.RawQuery = q.Encode()
	return u.String(), nil
}
