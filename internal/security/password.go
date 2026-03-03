package security

import (
	"aurora/internal/errorx"
	"crypto/rand"
	"crypto/subtle"
	"encoding/base64"
	"errors"
	"fmt"
	"strings"

	"golang.org/x/crypto/argon2"
)

// Argon2id parameters (reasonable production defaults).
// You can tune these based on your server capacity.
type Params struct {
	Memory      uint32 // in KiB
	Time        uint32 // iterations
	Parallelism uint8
	SaltLength  uint32
	KeyLength   uint32
}

func DefaultParams() Params {
	return Params{
		Memory:      64 * 1024, // 64 MiB
		Time:        3,
		Parallelism: 2,
		SaltLength:  16,
		KeyLength:   32,
	}
}

// HashPassword generates an encoded hash string for storage.
// Format:
// $argon2id$v=19$m=65536,t=3,p=2$<salt_b64>$<hash_b64>
func HashPassword(password string) (string, error) {
	return HashPasswordWithParams(password, DefaultParams())
}

func HashPasswordWithParams(password string, p Params) (string, error) {
	if password == "" {
		return "", errors.New("password is empty")
	}

	salt := make([]byte, p.SaltLength)
	if _, err := rand.Read(salt); err != nil {
		return "", fmt.Errorf("read salt: %w", err)
	}

	key := argon2.IDKey([]byte(password), salt, p.Time, p.Memory, p.Parallelism, p.KeyLength)

	b64Salt := base64.RawStdEncoding.EncodeToString(salt)
	b64Key := base64.RawStdEncoding.EncodeToString(key)

	encoded := fmt.Sprintf("$argon2id$v=19$m=%d,t=%d,p=%d$%s$%s",
		p.Memory, p.Time, p.Parallelism, b64Salt, b64Key,
	)
	return encoded, nil
}

// ComparePassword verifies the password against the stored encoded hash.
// Returns nil if match; otherwise ErrPasswordMismatch or parsing error.
func ComparePassword(storedHash, password string) error {
	if storedHash == "" {
		return errorx.ErrInvalidHashFormat
	}
	if password == "" {
		// still do parse to avoid behavior differences? up to you.
		return errorx.ErrPasswordMismatch
	}

	p, salt, hash, err := decodeHash(storedHash)
	if err != nil {
		return err
	}

	otherKey := argon2.IDKey([]byte(password), salt, p.Time, p.Memory, p.Parallelism, uint32(len(hash)))

	// constant-time compare
	if subtle.ConstantTimeCompare(hash, otherKey) != 1 {
		return errorx.ErrPasswordMismatch
	}
	return nil
}

func decodeHash(encoded string) (Params, []byte, []byte, error) {
	// Expected: $argon2id$v=19$m=...,t=...,p=...$salt$hash
	parts := strings.Split(encoded, "$")
	// Split("$...") yields leading empty string at index 0
	if len(parts) != 6 || parts[1] != "argon2id" {
		return Params{}, nil, nil, errorx.ErrInvalidHashFormat
	}

	// parts[2] = v=19 (we accept only v=19)
	if parts[2] != "v=19" {
		return Params{}, nil, nil, errorx.ErrInvalidHashFormat
	}

	var p Params
	// parts[3] = m=65536,t=3,p=2
	_, err := fmt.Sscanf(parts[3], "m=%d,t=%d,p=%d", &p.Memory, &p.Time, &p.Parallelism)
	if err != nil {
		return Params{}, nil, nil, errorx.ErrInvalidHashFormat
	}

	salt, err := base64.RawStdEncoding.DecodeString(parts[4])
	if err != nil || len(salt) == 0 {
		return Params{}, nil, nil, errorx.ErrInvalidHashFormat
	}

	hash, err := base64.RawStdEncoding.DecodeString(parts[5])
	if err != nil || len(hash) == 0 {
		return Params{}, nil, nil, errorx.ErrInvalidHashFormat
	}

	p.SaltLength = uint32(len(salt))
	p.KeyLength = uint32(len(hash))

	return p, salt, hash, nil
}
