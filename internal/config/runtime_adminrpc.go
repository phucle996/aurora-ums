package config

import (
	"fmt"
	"strconv"
	"strings"
	"time"
)

type UMSRuntimeBootstrap struct {
	App   UMSRuntimeBootstrapApp   `json:"app"`
	PSQL  UMSRuntimeBootstrapPSQL  `json:"psql"`
	Redis UMSRuntimeBootstrapRedis `json:"redis"`
	Token UMSRuntimeBootstrapToken `json:"token"`
}

type UMSRuntimeBootstrapApp struct {
	TimeZone string `json:"timezone"`
	LogLevel string `json:"log_level"`
	Port     int    `json:"port"`
}

type UMSRuntimeBootstrapPSQL struct {
	URL     string `json:"url"`
	SSLMode string `json:"ssl_mode"`
	Schema  string `json:"schema"`
}

type UMSRuntimeBootstrapRedis struct {
	Addr               string `json:"addr"`
	Username           string `json:"username"`
	Password           string `json:"password"`
	DB                 string `json:"db"`
	UseTLS             string `json:"use_tls"`
	CA                 string `json:"ca"`
	ClientKey          string `json:"client_key"`
	ClientCert         string `json:"client_cert"`
	InsecureSkipVerify string `json:"insecure_skip_verify"`
}

type UMSRuntimeBootstrapToken struct {
	AccessTTL  string `json:"access_ttl"`
	RefreshTTL string `json:"refresh_ttl"`
	DeviceTTL  string `json:"device_ttl"`
	OttTTL     string `json:"ott_ttl"`
}

func (cfg *Config) ApplyRuntimeBootstrap(bootstrap UMSRuntimeBootstrap) error {
	if cfg == nil {
		return fmt.Errorf("config is nil")
	}
	if strings.TrimSpace(bootstrap.App.TimeZone) == "" {
		return fmt.Errorf("missing runtime config: app.timezone")
	}
	if strings.TrimSpace(bootstrap.PSQL.URL) == "" {
		return fmt.Errorf("missing runtime config: psql.url")
	}
	if strings.TrimSpace(bootstrap.PSQL.SSLMode) == "" {
		return fmt.Errorf("missing runtime config: psql.ssl_mode")
	}
	if strings.TrimSpace(bootstrap.PSQL.Schema) == "" {
		return fmt.Errorf("missing runtime config: psql.schema")
	}
	if strings.TrimSpace(bootstrap.Redis.Addr) == "" {
		return fmt.Errorf("missing runtime config: redis.addr")
	}
	if strings.TrimSpace(bootstrap.Redis.DB) == "" {
		return fmt.Errorf("missing runtime config: redis.db")
	}
	if strings.TrimSpace(bootstrap.Redis.UseTLS) == "" {
		return fmt.Errorf("missing runtime config: redis.use_tls")
	}
	if strings.TrimSpace(bootstrap.Redis.InsecureSkipVerify) == "" {
		return fmt.Errorf("missing runtime config: redis.insecure_skip_verify")
	}
	if strings.TrimSpace(bootstrap.Token.AccessTTL) == "" {
		return fmt.Errorf("missing runtime config: token.access_ttl")
	}
	if strings.TrimSpace(bootstrap.Token.RefreshTTL) == "" {
		return fmt.Errorf("missing runtime config: token.refresh_ttl")
	}
	if strings.TrimSpace(bootstrap.Token.DeviceTTL) == "" {
		return fmt.Errorf("missing runtime config: token.device_ttl")
	}
	if strings.TrimSpace(bootstrap.Token.OttTTL) == "" {
		return fmt.Errorf("missing runtime config: token.ott_ttl")
	}

	cfg.App.TimeZone = strings.TrimSpace(bootstrap.App.TimeZone)
	if trimmed := strings.TrimSpace(bootstrap.App.LogLevel); trimmed != "" {
		cfg.App.LogLV = trimmed
	}
	if bootstrap.App.Port > 0 {
		cfg.App.Port = bootstrap.App.Port
	}

	cfg.Psql.DbURL = strings.TrimSpace(bootstrap.PSQL.URL)
	cfg.Psql.SslMode = strings.TrimSpace(bootstrap.PSQL.SSLMode)
	cfg.Psql.Schema = strings.TrimSpace(bootstrap.PSQL.Schema)

	cfg.Redis.Addr = strings.TrimSpace(bootstrap.Redis.Addr)
	cfg.Redis.Username = strings.TrimSpace(bootstrap.Redis.Username)
	cfg.Redis.Password = strings.TrimSpace(bootstrap.Redis.Password)

	redisDB, err := strconv.Atoi(strings.TrimSpace(bootstrap.Redis.DB))
	if err != nil {
		return fmt.Errorf("invalid runtime config: redis.db")
	}
	cfg.Redis.DB = redisDB

	cfg.Redis.UseTLS = parseBootstrapBool(bootstrap.Redis.UseTLS, cfg.Redis.UseTLS)
	cfg.Redis.CA = strings.TrimSpace(bootstrap.Redis.CA)
	cfg.Redis.ClientKey = strings.TrimSpace(bootstrap.Redis.ClientKey)
	cfg.Redis.ClientCert = strings.TrimSpace(bootstrap.Redis.ClientCert)
	cfg.Redis.InsecureSkipVerify = parseBootstrapBool(bootstrap.Redis.InsecureSkipVerify, cfg.Redis.InsecureSkipVerify)

	cfg.Token.AccessTTL, err = time.ParseDuration(strings.TrimSpace(bootstrap.Token.AccessTTL))
	if err != nil {
		return fmt.Errorf("invalid runtime config: token.access_ttl")
	}
	cfg.Token.RefreshTTL, err = time.ParseDuration(strings.TrimSpace(bootstrap.Token.RefreshTTL))
	if err != nil {
		return fmt.Errorf("invalid runtime config: token.refresh_ttl")
	}
	cfg.Token.DeviceTTL, err = time.ParseDuration(strings.TrimSpace(bootstrap.Token.DeviceTTL))
	if err != nil {
		return fmt.Errorf("invalid runtime config: token.device_ttl")
	}
	cfg.Token.OttTTL, err = time.ParseDuration(strings.TrimSpace(bootstrap.Token.OttTTL))
	if err != nil {
		return fmt.Errorf("invalid runtime config: token.ott_ttl")
	}

	return nil
}

func parseBootstrapBool(raw string, fallback bool) bool {
	switch strings.ToLower(strings.TrimSpace(raw)) {
	case "1", "true", "yes", "y", "on":
		return true
	case "0", "false", "no", "n", "off":
		return false
	default:
		return fallback
	}
}
