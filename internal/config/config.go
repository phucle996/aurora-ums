package config

import (
	"aurora/internal/cache"
	"log"
	"time"

	"github.com/joho/godotenv"
)

type AppCfg struct {
	Name     string
	Host     string
	Port     int
	LogLV    string
	TimeZone string
}

type PsqlCfg struct {
	DbURL      string
	Schema     string
	SslMode    string
	CA         string //  file path
	ClientKey  string // file path
	ClientCert string // file path
}

type RedisCfg struct {
	Addr               string
	Username           string
	Password           string
	DB                 int
	UseTLS             bool
	CA                 string // file path
	ClientKey          string // file path
	ClientCert         string // file path
	InsecureSkipVerify bool
}

type EtcdCfg struct {
	Endpoints []string

	AutoSyncInterval     time.Duration
	DialTimeout          time.Duration
	DialKeepAliveTime    time.Duration
	DialKeepAliveTimeout time.Duration

	Username string
	Password string

	UseTLS             bool
	CA                 string // file path
	ClientKey          string // file path
	ClientCert         string // file path
	ServerName         string
	InsecureSkipVerify bool

	PermitWithoutStream bool
	RejectOldCluster    bool
	MaxCallSendMsgSize  int
	MaxCallRecvMsgSize  int

	SASLEnable    bool
	SASLMechanism string
	SASLUsername  string
	SASLPassword  string
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

type TokenSecretSyncCfg struct {
	Enabled          bool
	Prefix           string
	BootstrapTimeout time.Duration
}

type Config struct {
	App             AppCfg
	Psql            PsqlCfg
	Redis           RedisCfg
	Etcd            EtcdCfg
	Token           TokenCfg
	TokenSecretSync TokenSecretSyncCfg
	Cors            CorsCfg
}

func LoadConfig() *Config {
	// load .env file (ignore if missing, can use env vars directly)
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, reading environment variables directly")
	}

	return &Config{
		App: AppCfg{
			Name:     "Aurora Cloud",
			Host:     getEnv("APP_HOST", ""),
			Port:     getEnvAsInt("APP_PORT", 3005),
			LogLV:    getEnv("APP_LOG_LEVEL", ""),
			TimeZone: getEnv("APP_TIMEZONE", "UTC"),
		},
		Psql: PsqlCfg{
			DbURL:      getEnv("DATABASE_URL", "postgres://aurora:27012004@localhost:5432/aurora"),
			Schema:     getEnv("DB_SCHEMA", "ums"),
			SslMode:    getEnv("DB_SSLMODE", "disable"),
			CA:         getEnv("DB_SSLROOTCERT", ""),
			ClientKey:  getEnv("DB_SSLKEY", ""),
			ClientCert: getEnv("DB_SSLCERT", ""),
		},
		Redis: RedisCfg{
			Addr:               getEnv("REDIS_ADDR", "localhost:6379"),
			Username:           getEnv("REDIS_USERNAME", ""),
			Password:           getEnv("REDIS_PASSWORD", ""),
			DB:                 getEnvAsInt("REDIS_DB", 0),
			UseTLS:             getEnvAsBool("REDIS_TLS", false),
			CA:                 getEnv("REDIS_TLS_CA", ""),
			ClientKey:          getEnv("REDIS_TLS_KEY", ""),
			ClientCert:         getEnv("REDIS_TLS_CERT", ""),
			InsecureSkipVerify: getEnvAsBool("REDIS_TLS_INSECURE", false),
		},
		Etcd: EtcdCfg{
			Endpoints:            getEnvAsSlice("ETCD_ENDPOINTS", []string{"localhost:2379"}),
			AutoSyncInterval:     getEnvAsDuration("ETCD_AUTO_SYNC_INTERVAL", 5*time.Minute),
			DialTimeout:          getEnvAsDuration("ETCD_DIAL_TIMEOUT", 5*time.Second),
			DialKeepAliveTime:    getEnvAsDuration("ETCD_DIAL_KEEPALIVE_TIME", 30*time.Second),
			DialKeepAliveTimeout: getEnvAsDuration("ETCD_DIAL_KEEPALIVE_TIMEOUT", 10*time.Second),

			Username: getEnv("ETCD_USERNAME", ""),
			Password: getEnv("ETCD_PASSWORD", ""),

			UseTLS:             getEnvAsBool("ETCD_TLS", false),
			CA:                 getEnv("ETCD_TLS_CA", ""),
			ClientKey:          getEnv("ETCD_TLS_KEY", ""),
			ClientCert:         getEnv("ETCD_TLS_CERT", ""),
			ServerName:         getEnv("ETCD_TLS_SERVER_NAME", ""),
			InsecureSkipVerify: getEnvAsBool("ETCD_TLS_INSECURE", false),

			PermitWithoutStream: getEnvAsBool("ETCD_PERMIT_WITHOUT_STREAM", false),
			RejectOldCluster:    getEnvAsBool("ETCD_REJECT_OLD_CLUSTER", false),
			MaxCallSendMsgSize:  getEnvAsInt("ETCD_MAX_CALL_SEND_MSG_SIZE", 2*1024*1024),
			MaxCallRecvMsgSize:  getEnvAsInt("ETCD_MAX_CALL_RECV_MSG_SIZE", 2*1024*1024),

			SASLEnable:    getEnvAsBool("ETCD_SASL_ENABLE", false),
			SASLMechanism: getEnv("ETCD_SASL_MECHANISM", "PLAIN"),
			SASLUsername:  getEnv("ETCD_SASL_USERNAME", ""),
			SASLPassword:  getEnv("ETCD_SASL_PASSWORD", ""),
		},
		Token: TokenCfg{
			AccessTTL:  15 * time.Minute,
			RefreshTTL: 168 * time.Hour,
			OttTTL:     15 * time.Minute,
			DeviceTTL:  15 * time.Minute,
			Secrets:    cache.NewTokenSecretCache(),
		},
		TokenSecretSync: TokenSecretSyncCfg{
			Enabled:          getEnvAsBool("TOKEN_SECRET_SYNC_ENABLED", true),
			Prefix:           getEnv("TOKEN_SECRET_SYNC_PREFIX", "/admin/token-secret"),
			BootstrapTimeout: getEnvAsDuration("TOKEN_SECRET_SYNC_BOOTSTRAP_TIMEOUT", 5*time.Second),
		},
		Cors: CorsCfg{
			AllowOrigins: getEnvAsSlice(
				"CORS_ALLOW_ORIGINS",
				[]string{
					"http://localhost:5173",
					"http://127.0.0.1:5173",
					"http://localhost:8080",
					"http://127.0.0.1:8080",
				},
			),

			AllowMethods: getEnvAsSlice(
				"CORS_ALLOW_METHODS",
				[]string{"GET", "POST", "PUT", "PATCH", "DELETE", "HEAD", "OPTIONS"},
			),
			AllowHeaders: getEnvAsSlice(
				"CORS_ALLOW_HEADERS",
				[]string{"Origin", "Content-Type", "Accept", "Authorization"},
			),
			ExposeHeaders:    getEnvAsSlice("CORS_EXPOSE_HEADERS", []string{}),
			AllowCredentials: getEnvAsBool("CORS_ALLOW_CREDENTIALS", true),
			MaxAge:           getEnvAsDuration("CORS_MAX_AGE", 12*time.Hour),
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

func (c *TokenCfg) ValidateEtcdBackedSecrets() error {
	if c == nil || c.Secrets == nil {
		return cache.ErrTokenSecretCacheNil
	}
	return c.Secrets.Validate()
}
