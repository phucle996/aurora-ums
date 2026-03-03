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

	clientv3 "go.etcd.io/etcd/client/v3"
)

const (
	adminTokenKindAccess  = "access_jwt"
	adminTokenKindRefresh = "refresh_jwt"
	adminTokenKindDevice  = "device_token"
)

type tokenSecretSync struct {
	client *clientv3.Client
	token  *config.TokenCfg
	prefix string
}

type tokenSecretRecord struct {
	Secret        string `json:"secret"`
	RotatedAtUnix int64  `json:"rotated_at_unix"`
}

func newTokenSecretSync(client *clientv3.Client, token *config.TokenCfg, prefix string) *tokenSecretSync {
	prefix = strings.TrimSpace(prefix)
	if prefix == "" {
		prefix = "/admin/token-secret"
	}
	return &tokenSecretSync{
		client: client,
		token:  token,
		prefix: prefix,
	}
}

func (s *tokenSecretSync) Bootstrap(ctx context.Context) error {
	if s == nil || s.client == nil || s.token == nil {
		return fmt.Errorf("token secret sync is not initialized")
	}
	for _, kind := range []string{adminTokenKindAccess, adminTokenKindRefresh, adminTokenKindDevice} {
		if err := s.syncKind(ctx, kind); err != nil {
			return err
		}
	}
	return nil
}

func (s *tokenSecretSync) Run(ctx context.Context) {
	if s == nil || s.client == nil || s.token == nil {
		return
	}

	watchPrefix := s.prefix + "/"
	wch := s.client.Watch(ctx, watchPrefix, clientv3.WithPrefix())

	for {
		select {
		case <-ctx.Done():
			return
		case wr, ok := <-wch:
			if !ok {
				logger.SysWarn("token.secret.sync", "watch channel closed")
				time.Sleep(500 * time.Millisecond)
				wch = s.client.Watch(ctx, watchPrefix, clientv3.WithPrefix())
				continue
			}
			if err := wr.Err(); err != nil {
				logger.SysWarn("token.secret.sync", "watch error: %v", err)
				continue
			}

			for _, ev := range wr.Events {
				key := strings.TrimSpace(string(ev.Kv.Key))
				if !strings.HasSuffix(key, "/current_version") {
					continue
				}
				kind, ok := s.kindFromCurrentVersionKey(key)
				if !ok {
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
}

func (s *tokenSecretSync) syncKind(ctx context.Context, kind string) error {
	currentVersionResp, err := s.client.Get(ctx, s.currentVersionKey(kind))
	if err != nil {
		return err
	}
	if len(currentVersionResp.Kvs) == 0 {
		return fmt.Errorf("missing current version for kind=%s", kind)
	}

	versionRaw := strings.TrimSpace(string(currentVersionResp.Kvs[0].Value))
	version, err := strconv.ParseInt(versionRaw, 10, 64)
	if err != nil || version <= 0 {
		return fmt.Errorf("invalid current version for kind=%s value=%q", kind, versionRaw)
	}

	recordResp, err := s.client.Get(ctx, s.versionKey(kind, version))
	if err != nil {
		return err
	}
	if len(recordResp.Kvs) == 0 {
		return fmt.Errorf("missing secret record kind=%s version=%d", kind, version)
	}

	recordRaw := strings.TrimSpace(string(recordResp.Kvs[0].Value))
	var rec tokenSecretRecord
	if err := json.Unmarshal([]byte(recordRaw), &rec); err != nil {
		return fmt.Errorf("decode token secret record kind=%s: %w", kind, err)
	}
	rec.Secret = strings.TrimSpace(rec.Secret)
	if rec.Secret == "" {
		return fmt.Errorf("empty secret for kind=%s version=%d", kind, version)
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

	logger.SysInfo("token.secret.sync", "updated token secret kind=%s version=%d", kind, version)
	return nil
}

func (s *tokenSecretSync) currentVersionKey(kind string) string {
	return fmt.Sprintf("%s/%s/current_version", s.prefix, strings.TrimSpace(kind))
}

func (s *tokenSecretSync) versionKey(kind string, version int64) string {
	return fmt.Sprintf("%s/%s/v/%d", s.prefix, strings.TrimSpace(kind), version)
}

func (s *tokenSecretSync) kindFromCurrentVersionKey(key string) (string, bool) {
	trimmed := strings.TrimSpace(strings.TrimPrefix(key, s.prefix+"/"))
	parts := strings.Split(trimmed, "/")
	if len(parts) < 2 {
		return "", false
	}
	if parts[len(parts)-1] != "current_version" {
		return "", false
	}
	kind := strings.TrimSpace(parts[0])
	switch kind {
	case adminTokenKindAccess, adminTokenKindRefresh, adminTokenKindDevice:
		return kind, true
	default:
		return "", false
	}
}
