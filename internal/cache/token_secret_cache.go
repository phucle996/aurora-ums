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
	AccessSecret  string
	RefreshSecret string
	DeviceSecret  string
	OttSecret     string

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
	c.AccessSecret = trimmed
	// Keep OTT secret derived from access secret so UMS does not use env for OTT.
	c.OttSecret = deriveOTTSecret(trimmed)
	c.mu.Unlock()
}

func (c *TokenSecretCache) SetRefreshSecret(v string) {
	if c == nil {
		return
	}
	trimmed := strings.TrimSpace(v)

	c.mu.Lock()
	c.RefreshSecret = trimmed
	c.mu.Unlock()
}

func (c *TokenSecretCache) SetDeviceSecret(v string) {
	if c == nil {
		return
	}
	trimmed := strings.TrimSpace(v)

	c.mu.Lock()
	c.DeviceSecret = trimmed
	c.mu.Unlock()
}

func (c *TokenSecretCache) GetAccessSecret() string {
	if c == nil {
		return ""
	}
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.AccessSecret
}

func (c *TokenSecretCache) GetRefreshSecret() string {
	if c == nil {
		return ""
	}
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.RefreshSecret
}

func (c *TokenSecretCache) GetDeviceSecret() string {
	if c == nil {
		return ""
	}
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.DeviceSecret
}

func (c *TokenSecretCache) GetOttSecret() string {
	if c == nil {
		return ""
	}
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.OttSecret
}

func (c *TokenSecretCache) Validate() error {
	if c == nil {
		return ErrTokenSecretCacheNil
	}
	c.mu.RLock()
	defer c.mu.RUnlock()

	if strings.TrimSpace(c.AccessSecret) == "" {
		return fmt.Errorf("missing access token secret from etcd")
	}
	if strings.TrimSpace(c.RefreshSecret) == "" {
		return fmt.Errorf("missing refresh token secret from etcd")
	}
	if strings.TrimSpace(c.DeviceSecret) == "" {
		return fmt.Errorf("missing device token secret from etcd")
	}
	if strings.TrimSpace(c.OttSecret) == "" {
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
