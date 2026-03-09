package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

var requiredUMSRuntimeKeys = []string{
	"app/timezone",
	"app/log_level",
	"app/port",
	"postgresql/url",
	"postgresql/sslmode",
	"postgresql/schema",
	"redis/addr",
	"redis/username",
	"redis/password",
	"redis/db",
	"redis/use_tls",
	"redis/ca",
	"redis/client_key",
	"redis/client_cert",
	"redis/insecure_skip_verify",
	"token_ttl/access_ttl",
	"token_ttl/refresh_ttl",
	"token_ttl/device_ttl",
	"token_ttl/ott_ttl",
	"cors/allow_origins",
	"cors/allow_methods",
	"cors/allow_headers",
	"cors/expose_headers",
	"cors/allow_credentials",
	"cors/max_age",
	"tls/ca_pem",
	"tls/client_cert_pem",
	"tls/client_key_pem",
}

func (cfg *Config) ApplyRuntimeValues(values map[string]string) error {
	if cfg == nil {
		return fmt.Errorf("config is nil")
	}
	if len(values) == 0 {
		return fmt.Errorf("runtime values are empty")
	}
	if err := validateRequiredRuntimeKeys(values); err != nil {
		return err
	}

	required := func(key string) (string, error) {
		v := strings.TrimSpace(values[key])
		if v == "" {
			return "", fmt.Errorf("missing runtime key: %s", key)
		}
		return v, nil
	}
	optional := func(key, def string) string {
		v := strings.TrimSpace(values[key])
		if v == "" {
			return def
		}
		return v
	}

	var err error
	if cfg.App.TimeZone, err = required("app/timezone"); err != nil {
		return err
	}
	cfg.App.LogLV = optional("app/log_level", cfg.App.LogLV)
	if appPort, err := strconv.Atoi(optional("app/port", strconv.Itoa(cfg.App.Port))); err == nil && appPort > 0 {
		cfg.App.Port = appPort
	}

	if cfg.Psql.DbURL, err = required("postgresql/url"); err != nil {
		return err
	}
	cfg.Psql.SslMode = optional("postgresql/sslmode", cfg.Psql.SslMode)
	cfg.Psql.Schema = optional("postgresql/schema", cfg.Psql.Schema)

	if cfg.Redis.Addr, err = required("redis/addr"); err != nil {
		return err
	}
	cfg.Redis.Username = optional("redis/username", "")
	cfg.Redis.Password = optional("redis/password", "")
	if redisDB, parseErr := strconv.Atoi(optional("redis/db", strconv.Itoa(cfg.Redis.DB))); parseErr == nil {
		cfg.Redis.DB = redisDB
	}
	cfg.Redis.UseTLS = parseBoolWithDefault(optional("redis/use_tls", ""), cfg.Redis.UseTLS)
	cfg.Redis.CA = optional("redis/ca", "")
	cfg.Redis.ClientKey = optional("redis/client_key", "")
	cfg.Redis.ClientCert = optional("redis/client_cert", "")
	cfg.Redis.InsecureSkipVerify = parseBoolWithDefault(optional("redis/insecure_skip_verify", ""), cfg.Redis.InsecureSkipVerify)

	cfg.Token.AccessTTL = parseDurationWithDefault(optional("token_ttl/access_ttl", ""), cfg.Token.AccessTTL)
	cfg.Token.RefreshTTL = parseDurationWithDefault(optional("token_ttl/refresh_ttl", ""), cfg.Token.RefreshTTL)
	cfg.Token.DeviceTTL = parseDurationWithDefault(optional("token_ttl/device_ttl", ""), cfg.Token.DeviceTTL)
	cfg.Token.OttTTL = parseDurationWithDefault(optional("token_ttl/ott_ttl", ""), cfg.Token.OttTTL)

	cfg.Cors.AllowOrigins = parseStringSliceWithFallback(optional("cors/allow_origins", ""), cfg.Cors.AllowOrigins)
	cfg.Cors.AllowMethods = parseStringSliceWithFallback(optional("cors/allow_methods", ""), cfg.Cors.AllowMethods)
	cfg.Cors.AllowHeaders = parseStringSliceWithFallback(optional("cors/allow_headers", ""), cfg.Cors.AllowHeaders)
	cfg.Cors.ExposeHeaders = parseStringSliceWithFallback(optional("cors/expose_headers", ""), cfg.Cors.ExposeHeaders)
	cfg.Cors.AllowCredentials = parseBoolWithDefault(optional("cors/allow_credentials", ""), cfg.Cors.AllowCredentials)
	cfg.Cors.MaxAge = parseDurationWithDefault(optional("cors/max_age", ""), cfg.Cors.MaxAge)

	if err := applyBootstrapTLSFiles(
		values,
		cfg.AdminRPC.CAPath,
		cfg.AdminRPC.ClientCert,
		cfg.AdminRPC.ClientKey,
	); err != nil {
		return err
	}
	return nil
}

func applyBootstrapTLSFiles(values map[string]string, caPath string, certPath string, keyPath string) error {
	caPEM := strings.TrimSpace(values["tls/ca_pem"])
	certPEM := strings.TrimSpace(values["tls/client_cert_pem"])
	keyPEM := strings.TrimSpace(values["tls/client_key_pem"])
	if caPEM == "" || certPEM == "" || keyPEM == "" {
		return fmt.Errorf("missing runtime key: tls bundle")
	}
	for _, path := range []string{caPath, certPath, keyPath} {
		if strings.TrimSpace(path) == "" {
			return fmt.Errorf("tls path is empty")
		}
		dir := filepath.Dir(path)
		if err := os.MkdirAll(dir, 0o750); err != nil {
			return fmt.Errorf("create tls dir failed (%s): %w", dir, err)
		}
	}
	if err := os.WriteFile(caPath, []byte(caPEM), 0o600); err != nil {
		return fmt.Errorf("write tls ca failed: %w", err)
	}
	if err := os.WriteFile(certPath, []byte(certPEM), 0o600); err != nil {
		return fmt.Errorf("write tls cert failed: %w", err)
	}
	if err := os.WriteFile(keyPath, []byte(keyPEM), 0o600); err != nil {
		return fmt.Errorf("write tls key failed: %w", err)
	}
	return nil
}

func validateRequiredRuntimeKeys(values map[string]string) error {
	missing := make([]string, 0)
	for _, key := range requiredUMSRuntimeKeys {
		if _, ok := values[key]; !ok {
			missing = append(missing, key)
		}
	}
	if len(missing) > 0 {
		return fmt.Errorf("missing runtime keys: %s", strings.Join(missing, ", "))
	}
	return nil
}

func parseDurationWithDefault(raw string, def time.Duration) time.Duration {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return def
	}
	d, err := time.ParseDuration(raw)
	if err != nil {
		return def
	}
	return d
}

func parseBoolWithDefault(raw string, def bool) bool {
	raw = strings.TrimSpace(strings.ToLower(raw))
	if raw == "" {
		return def
	}
	switch raw {
	case "1", "true", "yes", "y", "on":
		return true
	case "0", "false", "no", "n", "off":
		return false
	default:
		return def
	}
}

func parseStringSliceWithFallback(raw string, fallback []string) []string {
	out := parseRuntimeStringSlice(raw)
	if len(out) == 0 {
		return append([]string{}, fallback...)
	}
	return out
}

func parseRuntimeStringSlice(raw string) []string {
	trimmed := strings.TrimSpace(raw)
	if trimmed == "" {
		return nil
	}
	if strings.HasPrefix(trimmed, "[") {
		var arr []string
		if err := json.Unmarshal([]byte(trimmed), &arr); err == nil {
			out := make([]string, 0, len(arr))
			for _, item := range arr {
				item = strings.TrimSpace(item)
				if item != "" {
					out = append(out, item)
				}
			}
			return out
		}
	}

	parts := strings.Split(trimmed, ",")
	out := make([]string, 0, len(parts))
	for _, part := range parts {
		item := strings.Trim(strings.TrimSpace(part), `"'[]`)
		if item != "" {
			out = append(out, item)
		}
	}
	return out
}
