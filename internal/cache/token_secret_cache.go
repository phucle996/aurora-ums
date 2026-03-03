package cache

import (
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"fmt"
	"strings"
	"sync"
)

var ErrTokenSecretCacheNil = errors.New("token secret cache is nil")

// TokenSecretCache keeps etcd-synced token secrets in-memory.
type TokenSecretCache struct {
	accessSecret  string
	refreshSecret string
	deviceSecret  string
	ottSecret     string

	mu sync.RWMutex
}

func NewTokenSecretCache() *TokenSecretCache {
	return &TokenSecretCache{}
}

func (c *TokenSecretCache) SetAccessSecret(v string) {
	if c == nil {
		return
	}
	trimmed := strings.TrimSpace(v)

	c.mu.Lock()
	c.accessSecret = trimmed
	// Keep OTT secret derived from access secret so UMS does not use env for OTT.
	c.ottSecret = deriveOTTSecret(trimmed)
	c.mu.Unlock()
}

func (c *TokenSecretCache) SetRefreshSecret(v string) {
	if c == nil {
		return
	}
	trimmed := strings.TrimSpace(v)

	c.mu.Lock()
	c.refreshSecret = trimmed
	c.mu.Unlock()
}

func (c *TokenSecretCache) SetDeviceSecret(v string) {
	if c == nil {
		return
	}
	trimmed := strings.TrimSpace(v)

	c.mu.Lock()
	c.deviceSecret = trimmed
	c.mu.Unlock()
}

func (c *TokenSecretCache) GetAccessSecret() string {
	if c == nil {
		return ""
	}
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.accessSecret
}

func (c *TokenSecretCache) GetRefreshSecret() string {
	if c == nil {
		return ""
	}
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.refreshSecret
}

func (c *TokenSecretCache) GetDeviceSecret() string {
	if c == nil {
		return ""
	}
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.deviceSecret
}

func (c *TokenSecretCache) GetOttSecret() string {
	if c == nil {
		return ""
	}
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.ottSecret
}

func (c *TokenSecretCache) Validate() error {
	if c == nil {
		return ErrTokenSecretCacheNil
	}
	c.mu.RLock()
	defer c.mu.RUnlock()

	if strings.TrimSpace(c.accessSecret) == "" {
		return fmt.Errorf("missing access token secret from etcd")
	}
	if strings.TrimSpace(c.refreshSecret) == "" {
		return fmt.Errorf("missing refresh token secret from etcd")
	}
	if strings.TrimSpace(c.deviceSecret) == "" {
		return fmt.Errorf("missing device token secret from etcd")
	}
	if strings.TrimSpace(c.ottSecret) == "" {
		return fmt.Errorf("missing derived ott token secret")
	}
	return nil
}

func deriveOTTSecret(accessSecret string) string {
	trimmed := strings.TrimSpace(accessSecret)
	if trimmed == "" {
		return ""
	}
	sum := sha256.Sum256([]byte("ums:ott:" + trimmed))
	return base64.RawURLEncoding.EncodeToString(sum[:])
}
