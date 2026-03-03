package security

import (
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha1"
	"encoding/base32"
	"fmt"
	"net/url"
	"strings"
	"time"
)

func GenerateTOTPSecret() (string, error) {
	buf := make([]byte, 20)
	if _, err := rand.Read(buf); err != nil {
		return "", err
	}
	enc := base32.StdEncoding.WithPadding(base32.NoPadding)
	return enc.EncodeToString(buf), nil
}

func BuildOTPAuthURL(account, secret string) string {
	issuer := "Aurora"
	label := url.PathEscape(fmt.Sprintf("%s:%s", issuer, account))
	return fmt.Sprintf(
		"otpauth://totp/%s?secret=%s&issuer=%s",
		label,
		url.QueryEscape(secret),
		url.QueryEscape(issuer),
	)
}

func ValidateTOTP(secret, code string, allowedSkew int) bool {
	secret = normalizeOTPSecret(secret)
	code = strings.TrimSpace(code)
	if secret == "" || code == "" {
		return false
	}

	decoded, err := base32.StdEncoding.WithPadding(base32.NoPadding).DecodeString(secret)
	if err != nil {
		return false
	}

	if allowedSkew < 0 {
		allowedSkew = 0
	}

	now := time.Now().Unix()
	step := int64(30)
	counter := now / step

	for i := -allowedSkew; i <= allowedSkew; i++ {
		if generateTOTP(decoded, counter+int64(i)) == code {
			return true
		}
	}
	return false
}

func generateTOTP(secret []byte, counter int64) string {
	var buf [8]byte
	for i := 7; i >= 0; i-- {
		buf[i] = byte(counter & 0xff)
		counter >>= 8
	}

	mac := hmac.New(sha1.New, secret)
	mac.Write(buf[:])
	sum := mac.Sum(nil)

	offset := sum[len(sum)-1] & 0x0f
	bin := (int(sum[offset])&0x7f)<<24 |
		(int(sum[offset+1])&0xff)<<16 |
		(int(sum[offset+2])&0xff)<<8 |
		(int(sum[offset+3]) & 0xff)

	otp := bin % 1000000
	return fmt.Sprintf("%06d", otp)
}

func normalizeOTPSecret(secret string) string {
	secret = strings.ReplaceAll(secret, " ", "")
	secret = strings.ReplaceAll(secret, "-", "")
	return strings.ToUpper(secret)
}
