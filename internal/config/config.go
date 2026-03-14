package config

import (
	"aurora/internal/cache"
	"os"
	"strings"
	"time"
)

const (
	UMSTLSCertPath            = "/etc/aurora/certs/ums.crt"
	UMSTLSKeyPath             = "/etc/aurora/certs/ums.key"
	UMSTLSCAPath              = "/etc/aurora/certs/ca.crt"
	UMSAdminRPCClientCertPath = "/etc/aurora/certs/ums-adminrpc-client.crt"
	UMSAdminRPCClientKeyPath  = "/etc/aurora/certs/ums-adminrpc-client.key"
)

type AppCfg struct {
	Name        string
	Host        string
	Port        int
	LogLV       string
	TimeZone    string
	TLSCertPath string
	TLSKeyPath  string
	TLSCAPath   string
}

type PsqlCfg struct {
	DbURL      string
	Schema     string
	SslMode    string
	CA         string
	ClientKey  string
	ClientCert string
}

type RedisCfg struct {
	Addr               string
	Username           string
	Password           string
	DB                 int
	UseTLS             bool
	CA                 string
	ClientKey          string
	ClientCert         string
	InsecureSkipVerify bool
}

type AdminRPCCfg struct {
	Endpoint           string
	InsecureSkipVerify bool
	DialTimeout        time.Duration
	CAPath             string
	ClientCert         string
	ClientKey          string
	BootstrapToken     string
}

type TokenCfg struct {
	AccessTTL  time.Duration
	RefreshTTL time.Duration
	OttTTL     time.Duration
	DeviceTTL  time.Duration
	Secrets    *cache.TokenSecretCache
}

type CorsCfg struct {
	AllowOrigins     []string
	AllowMethods     []string
	AllowHeaders     []string
	ExposeHeaders    []string
	AllowCredentials bool
	MaxAge           time.Duration
}

type Config struct {
	App      AppCfg
	Psql     PsqlCfg
	Redis    RedisCfg
	AdminRPC AdminRPCCfg
	Token    TokenCfg
	Cors     CorsCfg
}

func LoadConfig() *Config {
	return &Config{
		App: AppCfg{
			Name:        "Aurora User Managment System",
			Host:        "",
			Port:        3005,
			LogLV:       "info",
			TimeZone:    "UTC",
			TLSCertPath: UMSTLSCertPath,
			TLSKeyPath:  UMSTLSKeyPath,
			TLSCAPath:   UMSTLSCAPath,
		},
		Psql: PsqlCfg{
			DbURL:      "",
			Schema:     "ums",
			SslMode:    "disable",
			CA:         "",
			ClientKey:  "",
			ClientCert: "",
		},
		Redis: RedisCfg{
			Addr:               "",
			Username:           "",
			Password:           "",
			DB:                 0,
			UseTLS:             false,
			CA:                 "",
			ClientKey:          "",
			ClientCert:         "",
			InsecureSkipVerify: false,
		},
		AdminRPC: AdminRPCCfg{
			Endpoint:       strings.TrimSpace(os.Getenv("ADMIN_RPC_ENDPOINT")),
			DialTimeout:    5 * time.Second,
			CAPath:         UMSTLSCAPath,
			ClientCert:     UMSAdminRPCClientCertPath,
			ClientKey:      UMSAdminRPCClientKeyPath,
			BootstrapToken: strings.TrimSpace(os.Getenv("ADMIN_RPC_BOOTSTRAP_TOKEN")),
		},
		Token: TokenCfg{
			AccessTTL:  15 * time.Minute,
			RefreshTTL: 168 * time.Hour,
			OttTTL:     15 * time.Minute,
			DeviceTTL:  15 * time.Minute,
			Secrets:    cache.NewTokenSecretCache(),
		},
		Cors: CorsCfg{
			AllowOrigins: []string{
				"http://localhost:5173",
				"http://127.0.0.1:5173",
				"http://localhost:8080",
				"http://127.0.0.1:8080",
			},
			AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE", "HEAD", "OPTIONS"},
			AllowHeaders:     []string{"Origin", "Content-Type", "Accept", "Authorization"},
			ExposeHeaders:    []string{},
			AllowCredentials: true,
			MaxAge:           12 * time.Hour,
		},
	}
}

func (c *TokenCfg) GetAccessSecret() string {
	if c == nil || c.Secrets == nil {
		return ""
	}
	return c.Secrets.GetAccessSecret()
}

func (c *TokenCfg) SetAccessSecret(v string) {
	if c == nil {
		return
	}
	if c.Secrets == nil {
		c.Secrets = cache.NewTokenSecretCache()
	}
	c.Secrets.SetAccessSecret(v)
}

func (c *TokenCfg) GetRefreshSecret() string {
	if c == nil || c.Secrets == nil {
		return ""
	}
	return c.Secrets.GetRefreshSecret()
}

func (c *TokenCfg) SetRefreshSecret(v string) {
	if c == nil {
		return
	}
	if c.Secrets == nil {
		c.Secrets = cache.NewTokenSecretCache()
	}
	c.Secrets.SetRefreshSecret(v)
}

func (c *TokenCfg) GetDeviceSecret() string {
	if c == nil || c.Secrets == nil {
		return ""
	}
	return c.Secrets.GetDeviceSecret()
}

func (c *TokenCfg) SetDeviceSecret(v string) {
	if c == nil {
		return
	}
	if c.Secrets == nil {
		c.Secrets = cache.NewTokenSecretCache()
	}
	c.Secrets.SetDeviceSecret(v)
}

func (c *TokenCfg) GetOttSecret() string {
	if c == nil || c.Secrets == nil {
		return ""
	}
	return c.Secrets.GetOttSecret()
}

func (c *TokenCfg) ValidateSecrets() error {
	if c == nil || c.Secrets == nil {
		return cache.ErrTokenSecretCacheNil
	}
	return c.Secrets.Validate()
}
