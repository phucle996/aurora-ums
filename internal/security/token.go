package security

import (
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"errors"
)

// GenerateRandomToken returns a cryptographically secure random token.
// length = số byte random (không phải độ dài string sau encode).
func GenerateToken(length int) (string, error) {
	if length <= 0 {
		return "", errors.New("token length must be positive")
	}

	buf := make([]byte, length)
	if _, err := rand.Read(buf); err != nil {
		return "", err
	}

	// URL-safe, không padding
	return base64.RawURLEncoding.EncodeToString(buf), nil
}

// HashToken creates a deterministic HMAC-SHA256 hash for storage/comparison.
func HashToken(token, secret string) (string, error) {
	if token == "" {
		return "", errors.New("token is empty")
	}
	if secret == "" {
		return "", errors.New("secret is empty")
	}
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write([]byte(token))
	return base64.RawURLEncoding.EncodeToString(mac.Sum(nil)), nil
}
