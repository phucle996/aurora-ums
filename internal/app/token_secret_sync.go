package app

import (
	"aurora/internal/config"
	"aurora/pkg/logger"
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/redis/go-redis/v9"
)

const (
	adminTokenKindAccess  = "access_jwt"
	adminTokenKindRefresh = "refresh_jwt"
	adminTokenKindDevice  = "device_token"
	tokenSecretCachePrefixDefault = "aurora:token-secret"
	tokenSecretInvalidateChannel  = "aurora:token-secret:invalidate"
	tokenSecretPollInterval       = 10 * time.Second
	tokenSecretBootstrapTimeout   = 5 * time.Second
)

type tokenSecretSync struct {
	redis             *redis.Client
	token             *config.TokenCfg
	cacheKeyPrefix    string
	invalidateChannel string
	pollInterval      time.Duration
}

type tokenSecretRecord struct {
	Kind          string `json:"kind"`
	Version       int64  `json:"version"`
	Secret        string `json:"secret"`
	RotatedAtUnix int64  `json:"rotated_at_unix"`
}

func newTokenSecretSync(
	redisClient *redis.Client,
	token *config.TokenCfg,
) *tokenSecretSync {
	return &tokenSecretSync{
		redis:             redisClient,
		token:             token,
		cacheKeyPrefix:    tokenSecretCachePrefixDefault,
		invalidateChannel: tokenSecretInvalidateChannel,
		pollInterval:      tokenSecretPollInterval,
	}
}

func (s *tokenSecretSync) Bootstrap(ctx context.Context) error {
	if s == nil || s.redis == nil || s.token == nil {
		return fmt.Errorf("token secret redis sync is not initialized")
	}
	for _, kind := range []string{adminTokenKindAccess, adminTokenKindRefresh, adminTokenKindDevice} {
		if err := s.syncKind(ctx, kind); err != nil {
			return err
		}
	}
	return nil
}

func (s *tokenSecretSync) Run(ctx context.Context) {
	if s == nil || s.redis == nil || s.token == nil {
		return
	}

	ticker := time.NewTicker(s.pollInterval)
	defer ticker.Stop()

	sub := s.redis.Subscribe(ctx, s.invalidateChannel)
	defer sub.Close()
	channel := sub.Channel()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			pollCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
			if err := s.Bootstrap(pollCtx); err != nil {
				logger.SysWarn("token.secret.sync", "poll sync failed: %v", err)
			}
			cancel()
		case msg, ok := <-channel:
			if !ok {
				logger.SysWarn("token.secret.sync", "redis invalidate channel closed")
				time.Sleep(500 * time.Millisecond)
				sub = s.redis.Subscribe(ctx, s.invalidateChannel)
				channel = sub.Channel()
				continue
			}
			kind, _, parseErr := parseInvalidateMessage(msg.Payload)
			if parseErr != nil {
				logger.SysWarn("token.secret.sync", "invalid invalidate message=%q", msg.Payload)
				continue
			}
			syncCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
			if err := s.syncKind(syncCtx, kind); err != nil {
				logger.SysWarn("token.secret.sync", "sync secret failed kind=%s err=%v", kind, err)
			}
			cancel()
		}
	}
}

func (s *tokenSecretSync) syncKind(ctx context.Context, kind string) error {
	key := s.secretKey(kind)
	raw, err := s.redis.Get(ctx, key).Result()
	if err != nil {
		return err
	}

	var rec tokenSecretRecord
	if err := json.Unmarshal([]byte(strings.TrimSpace(raw)), &rec); err != nil {
		return err
	}
	rec.Secret = strings.TrimSpace(rec.Secret)
	if rec.Secret == "" {
		return fmt.Errorf("empty secret for kind=%s", kind)
	}

	switch kind {
	case adminTokenKindAccess:
		s.token.SetAccessSecret(rec.Secret)
	case adminTokenKindRefresh:
		s.token.SetRefreshSecret(rec.Secret)
	case adminTokenKindDevice:
		s.token.SetDeviceSecret(rec.Secret)
	default:
		return fmt.Errorf("unknown token kind=%s", kind)
	}
	logger.SysInfo("token.secret.sync", "updated token secret kind=%s version=%d", kind, rec.Version)
	return nil
}

func (s *tokenSecretSync) secretKey(kind string) string {
	return strings.TrimRight(s.cacheKeyPrefix, ":") + ":" + strings.TrimSpace(kind)
}

func parseInvalidateMessage(raw string) (string, int64, error) {
	trimmed := strings.TrimSpace(raw)
	if trimmed == "" {
		return "", 0, fmt.Errorf("empty invalidate message")
	}
	kind, versionRaw, ok := strings.Cut(trimmed, ":")
	if !ok {
		return "", 0, fmt.Errorf("invalid invalidate format")
	}
	version, err := strconv.ParseInt(strings.TrimSpace(versionRaw), 10, 64)
	if err != nil {
		return "", 0, err
	}
	kind = strings.TrimSpace(kind)
	switch kind {
	case adminTokenKindAccess, adminTokenKindRefresh, adminTokenKindDevice:
		return kind, version, nil
	default:
		return "", 0, fmt.Errorf("unsupported kind=%s", kind)
	}
}
