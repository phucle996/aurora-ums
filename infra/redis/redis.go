package redisinfra

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"os"

	"aurora/internal/config"
	"aurora/pkg/logger"

	"github.com/redis/go-redis/v9"
)

// NewRedis creates a redis client with optional TLS and verifies connectivity via PING.
func NewRedis(ctx context.Context, cfg *config.RedisCfg) (*redis.Client, error) {
	const op = "system.redis.connect"

	tlsCfg, err := buildTLSConfig(cfg)
	if err != nil {
		logger.SysError(op, err, "failed to build TLS config")
		return nil, err
	}

	opt := &redis.Options{
		Addr:      cfg.Addr,
		Username:  cfg.Username,
		Password:  cfg.Password,
		DB:        cfg.DB,
		TLSConfig: tlsCfg,
	}

	client := redis.NewClient(opt)

	if err := client.Ping(ctx).Err(); err != nil {
		logger.SysError(op, err, "redis ping failed")
		return nil, fmt.Errorf("redis ping: %w", err)
	}

	logger.SysInfo(op, "connected to redis %s (tls=%v)", cfg.Addr, tlsCfg != nil)
	return client, nil
}

func buildTLSConfig(cfg *config.RedisCfg) (*tls.Config, error) {
	if !cfg.UseTLS && cfg.CA == "" && cfg.ClientCert == "" && cfg.ClientKey == "" {
		return nil, nil
	}

	tlsCfg := &tls.Config{
		InsecureSkipVerify: cfg.InsecureSkipVerify, //nolint:gosec // configurable for staging/debug
	}

	if cfg.CA != "" {
		rootCAs := x509.NewCertPool()
		caPem, err := os.ReadFile(cfg.CA)
		if err != nil {
			return nil, fmt.Errorf("read redis CA: %w", err)
		}
		if ok := rootCAs.AppendCertsFromPEM(caPem); !ok {
			return nil, fmt.Errorf("append redis CA failed")
		}
		tlsCfg.RootCAs = rootCAs
	}

	if cfg.ClientCert != "" && cfg.ClientKey != "" {
		cert, err := tls.LoadX509KeyPair(cfg.ClientCert, cfg.ClientKey)
		if err != nil {
			return nil, fmt.Errorf("load redis client cert/key: %w", err)
		}
		tlsCfg.Certificates = []tls.Certificate{cert}
	}

	return tlsCfg, nil
}
