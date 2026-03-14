package config

import (
	"testing"
	"time"
)

func TestApplyRuntimeBootstrap_UsesStructuredConfig(t *testing.T) {
	cfg := LoadConfig()
	tlsDir := t.TempDir()
	cfg.AdminRPC.CAPath = tlsDir + "/ca.crt"
	cfg.AdminRPC.ClientCert = tlsDir + "/client.crt"
	cfg.AdminRPC.ClientKey = tlsDir + "/client.key"

	origOrigins := append([]string{}, cfg.Cors.AllowOrigins...)
	origMaxAge := cfg.Cors.MaxAge

	err := cfg.ApplyRuntimeBootstrap(UMSRuntimeBootstrap{
		App: UMSRuntimeBootstrapApp{
			TimeZone: "Asia/Ho_Chi_Minh",
			LogLevel: "debug",
			Port:     3005,
		},
		PSQL: UMSRuntimeBootstrapPSQL{
			URL:     "postgres://user:pass@localhost:5432/ums",
			SSLMode: "disable",
			Schema:  "ums_test",
		},
		Redis: UMSRuntimeBootstrapRedis{
			Addr:               "localhost:6379",
			DB:                 "1",
			UseTLS:             "false",
			InsecureSkipVerify: "false",
		},
		Token: UMSRuntimeBootstrapToken{
			AccessTTL:  "15m",
			RefreshTTL: "168h",
			DeviceTTL:  "15m",
			OttTTL:     "15m",
		},
		TLS: UMSRuntimeBootstrapTLS{
			CAPEM:         "ca",
			ClientCertPEM: "cert",
			ClientKeyPEM:  "key",
		},
	})
	if err != nil {
		t.Fatalf("ApplyRuntimeBootstrap() error = %v", err)
	}

	if cfg.Psql.Schema != "ums_test" {
		t.Fatalf("expected schema to be updated, got %q", cfg.Psql.Schema)
	}
	if cfg.Redis.DB != 1 {
		t.Fatalf("expected redis db to be updated, got %d", cfg.Redis.DB)
	}
	if cfg.Token.AccessTTL != 15*time.Minute {
		t.Fatalf("expected access ttl to update, got %s", cfg.Token.AccessTTL)
	}
	if len(cfg.Cors.AllowOrigins) != len(origOrigins) {
		t.Fatalf("expected default cors origins to remain unchanged")
	}
	if cfg.Cors.MaxAge != origMaxAge {
		t.Fatalf("expected default cors max age to remain unchanged")
	}
}
