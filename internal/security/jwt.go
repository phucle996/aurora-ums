package security

import (
	"aurora/internal/errorx"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"
)

// GenerateJWT creates a HS256 JWT with standard claims.
// subject can be empty. extra claims are merged in (extra takes precedence).
func GenerateJWT(subject, secret string, ttl time.Duration, extra map[string]any) (string, error) {
	if secret == "" {
		return "", errors.New("missing jwt secret")
	}

	header := map[string]string{
		"alg": "HS256",
		"typ": "JWT",
	}

	now := time.Now().Unix()
	claims := map[string]any{
		"iat": now,
	}
	if ttl > 0 {
		claims["exp"] = now + int64(ttl.Seconds())
	}
	if subject != "" {
		claims["sub"] = subject
	}
	for k, v := range extra {
		claims[k] = v
	}

	headerJSON, err := json.Marshal(header)
	if err != nil {
		return "", fmt.Errorf("marshal header: %w", err)
	}
	claimsJSON, err := json.Marshal(claims)
	if err != nil {
		return "", fmt.Errorf("marshal claims: %w", err)
	}

	encodedHeader := base64.RawURLEncoding.EncodeToString(headerJSON)
	encodedClaims := base64.RawURLEncoding.EncodeToString(claimsJSON)
	signingInput := encodedHeader + "." + encodedClaims
	signature := signHS256(signingInput, secret)

	return signingInput + "." + signature, nil
}

// DecodeJWT verifies HS256 signature and validates time-based claims.
// Returns all claims if valid.
func DecodeJWT(token, secret string) (map[string]any, error) {
	parts := strings.Split(token, ".")
	if len(parts) != 3 {
		return nil, errorx.ErrTokenInvalid
	}
	headerSeg, payloadSeg, sigSeg := parts[0], parts[1], parts[2]

	signingInput := headerSeg + "." + payloadSeg
	expectedSig := signHS256(signingInput, secret)
	if !hmac.Equal([]byte(sigSeg), []byte(expectedSig)) {
		return nil, errorx.ErrTokenInvalid
	}

	headerBytes, err := base64.RawURLEncoding.DecodeString(headerSeg)
	if err != nil {
		return nil, errorx.ErrTokenInvalid
	}
	var header map[string]any
	if err := json.Unmarshal(headerBytes, &header); err != nil {
		return nil, errorx.ErrTokenInvalid
	}
	if alg, ok := header["alg"].(string); !ok || alg != "HS256" {
		return nil, errorx.ErrTokenInvalid
	}

	payloadBytes, err := base64.RawURLEncoding.DecodeString(payloadSeg)
	if err != nil {
		return nil, errorx.ErrTokenInvalid
	}
	var claims map[string]any
	if err := json.Unmarshal(payloadBytes, &claims); err != nil {
		return nil, errorx.ErrTokenInvalid
	}

	now := time.Now().Unix()
	if exp, ok := getNumericClaim(claims, "exp"); ok && exp > 0 && now > exp {
		return nil, errorx.ErrTokenExpired
	}
	if nbf, ok := getNumericClaim(claims, "nbf"); ok && nbf > 0 && now < nbf {
		return nil, errorx.ErrTokenNotActive
	}

	return claims, nil
}

// DecodeJWTAllowExpired verifies signature and returns claims even if exp is in the past.
func DecodeJWTAllowExpired(token, secret string) (map[string]any, error) {
	parts := strings.Split(token, ".")
	if len(parts) != 3 {
		return nil, errorx.ErrTokenInvalid
	}
	headerSeg, payloadSeg, sigSeg := parts[0], parts[1], parts[2]

	signingInput := headerSeg + "." + payloadSeg
	expectedSig := signHS256(signingInput, secret)
	if !hmac.Equal([]byte(sigSeg), []byte(expectedSig)) {
		return nil, errorx.ErrTokenInvalid
	}

	headerBytes, err := base64.RawURLEncoding.DecodeString(headerSeg)
	if err != nil {
		return nil, errorx.ErrTokenInvalid
	}
	var header map[string]any
	if err := json.Unmarshal(headerBytes, &header); err != nil {
		return nil, errorx.ErrTokenInvalid
	}
	if alg, ok := header["alg"].(string); !ok || alg != "HS256" {
		return nil, errorx.ErrTokenInvalid
	}

	payloadBytes, err := base64.RawURLEncoding.DecodeString(payloadSeg)
	if err != nil {
		return nil, errorx.ErrTokenInvalid
	}
	var claims map[string]any
	if err := json.Unmarshal(payloadBytes, &claims); err != nil {
		return nil, errorx.ErrTokenInvalid
	}

	now := time.Now().Unix()
	if nbf, ok := getNumericClaim(claims, "nbf"); ok && nbf > 0 && now < nbf {
		return nil, errorx.ErrTokenNotActive
	}

	return claims, nil
}

func signHS256(message, secret string) string {
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write([]byte(message))
	return base64.RawURLEncoding.EncodeToString(mac.Sum(nil))
}

func getNumericClaim(claims map[string]any, key string) (int64, bool) {
	val, ok := claims[key]
	if !ok {
		return 0, false
	}
	switch v := val.(type) {
	case float64:
		return int64(v), true
	case int64:
		return v, true
	case int:
		return int64(v), true
	case json.Number:
		i, err := v.Int64()
		if err != nil {
			return 0, false
		}
		return i, true
	default:
		return 0, false
	}
}
