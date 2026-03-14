package adminrpc

import (
	"aurora/internal/config"
	"context"
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"net"
	"net/url"
	"os"
	"strings"

	gogrpc "google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/protobuf/types/known/structpb"
)

const runtimeGetRuntimeBootstrapPath = "/admin.transport.runtime.v1.RuntimeService/GetRuntimeBootstrap"

func FetchUMSRuntimeBootstrap(ctx context.Context, cfg *config.AdminRPCCfg) (*config.UMSRuntimeBootstrap, error) {
	if cfg == nil {
		return nil, fmt.Errorf("admin rpc config is nil")
	}

	endpoint := strings.TrimSpace(cfg.Endpoint)
	if endpoint == "" {
		return nil, fmt.Errorf("admin rpc endpoint is empty")
	}
	target, serverName, err := normalizeAdminRPCTarget(endpoint)
	if err != nil {
		return nil, err
	}

	tlsCfg, err := buildTLSConfig(cfg, serverName)
	if err != nil {
		return nil, err
	}
	conn, err := gogrpc.NewClient(target, gogrpc.WithTransportCredentials(credentials.NewTLS(tlsCfg)))
	if err != nil {
		return nil, fmt.Errorf("dial admin rpc failed: %w", err)
	}
	defer conn.Close()

	req, err := structpb.NewStruct(map[string]any{
		"module_name": "ums",
		"config_keys": []any{
			"app.timezone",
			"app.log_level",
			"app.port",
			"psql.url",
			"psql.ssl_mode",
			"psql.schema",
			"redis.addr",
			"redis.username",
			"redis.password",
			"redis.db",
			"redis.use_tls",
			"redis.ca",
			"redis.client_key",
			"redis.client_cert",
			"redis.insecure_skip_verify",
			"token.access_ttl",
			"token.refresh_ttl",
			"token.device_ttl",
			"token.ott_ttl",
			"tls.ca_pem",
			"tls.client_cert_pem",
			"tls.client_key_pem",
		},
	})
	if err != nil {
		return nil, err
	}
	res := &structpb.Struct{}
	if err := conn.Invoke(ctx, runtimeGetRuntimeBootstrapPath, req, res); err != nil {
		return nil, fmt.Errorf("invoke admin runtime rpc failed: %w", err)
	}

	configField, ok := res.GetFields()["config"]
	if !ok || configField == nil || configField.GetStructValue() == nil {
		return nil, fmt.Errorf("runtime rpc response missing config")
	}
	root := configField.GetStructValue().GetFields()
	appField := root["app"]
	psqlField := root["psql"]
	redisField := root["redis"]
	tokenField := root["token"]
	tlsField := root["tls"]
	if appField == nil || appField.GetStructValue() == nil {
		return nil, fmt.Errorf("runtime rpc response missing config.app")
	}
	if psqlField == nil || psqlField.GetStructValue() == nil {
		return nil, fmt.Errorf("runtime rpc response missing config.psql")
	}
	if redisField == nil || redisField.GetStructValue() == nil {
		return nil, fmt.Errorf("runtime rpc response missing config.redis")
	}
	if tokenField == nil || tokenField.GetStructValue() == nil {
		return nil, fmt.Errorf("runtime rpc response missing config.token")
	}
	if tlsField == nil || tlsField.GetStructValue() == nil {
		return nil, fmt.Errorf("runtime rpc response missing config.tls")
	}

	app := appField.GetStructValue().GetFields()
	psql := psqlField.GetStructValue().GetFields()
	redis := redisField.GetStructValue().GetFields()
	token := tokenField.GetStructValue().GetFields()
	tlsBundle := tlsField.GetStructValue().GetFields()

	runtimeCfg := &config.UMSRuntimeBootstrap{
		App: config.UMSRuntimeBootstrapApp{
			TimeZone: readStructString(app, "timezone"),
			LogLevel: readStructString(app, "log_level"),
			Port:     readStructInt(app, "port"),
		},
		PSQL: config.UMSRuntimeBootstrapPSQL{
			URL:     readStructString(psql, "url"),
			SSLMode: readStructString(psql, "ssl_mode"),
			Schema:  readStructString(psql, "schema"),
		},
		Redis: config.UMSRuntimeBootstrapRedis{
			Addr:               readStructString(redis, "addr"),
			Username:           readStructString(redis, "username"),
			Password:           readStructString(redis, "password"),
			DB:                 readStructString(redis, "db"),
			UseTLS:             readStructString(redis, "use_tls"),
			CA:                 readStructString(redis, "ca"),
			ClientKey:          readStructString(redis, "client_key"),
			ClientCert:         readStructString(redis, "client_cert"),
			InsecureSkipVerify: readStructString(redis, "insecure_skip_verify"),
		},
		Token: config.UMSRuntimeBootstrapToken{
			AccessTTL:  readStructString(token, "access_ttl"),
			RefreshTTL: readStructString(token, "refresh_ttl"),
			DeviceTTL:  readStructString(token, "device_ttl"),
			OttTTL:     readStructString(token, "ott_ttl"),
		},
		TLS: config.UMSRuntimeBootstrapTLS{
			CAPEM:         readStructString(tlsBundle, "ca_pem"),
			ClientCertPEM: readStructString(tlsBundle, "client_cert_pem"),
			ClientKeyPEM:  readStructString(tlsBundle, "client_key_pem"),
		},
	}
	return runtimeCfg, nil
}

func buildTLSConfig(cfg *config.AdminRPCCfg, serverName string) (*tls.Config, error) {
	tlsCfg := &tls.Config{
		MinVersion: tls.VersionTLS12,
		ServerName: serverName,
	}

	caPath := strings.TrimSpace(cfg.CAPath)
	if caPath == "" {
		return nil, fmt.Errorf("admin rpc ca path is required for mTLS")
	}
	caPEM, err := os.ReadFile(caPath)
	if err != nil {
		return nil, fmt.Errorf("read admin rpc ca failed: %w", err)
	}
	pool := x509.NewCertPool()
	if ok := pool.AppendCertsFromPEM(caPEM); !ok {
		return nil, fmt.Errorf("invalid admin rpc ca")
	}
	tlsCfg.RootCAs = pool

	certPath := strings.TrimSpace(cfg.ClientCert)
	keyPath := strings.TrimSpace(cfg.ClientKey)
	if certPath == "" || keyPath == "" {
		return nil, fmt.Errorf("admin rpc client cert/key is required for mTLS")
	}
	pair, err := tls.LoadX509KeyPair(certPath, keyPath)
	if err != nil {
		return nil, fmt.Errorf("load admin rpc client cert/key failed: %w", err)
	}
	tlsCfg.Certificates = []tls.Certificate{pair}
	return tlsCfg, nil
}

func normalizeAdminRPCTarget(endpoint string) (target string, serverName string, err error) {
	raw := strings.TrimSpace(endpoint)
	if raw == "" {
		return "", "", fmt.Errorf("admin rpc endpoint is empty")
	}
	if strings.Contains(raw, "://") {
		parsed, parseErr := url.Parse(raw)
		if parseErr != nil {
			return "", "", fmt.Errorf("invalid admin rpc endpoint %q", endpoint)
		}
		switch strings.ToLower(strings.TrimSpace(parsed.Scheme)) {
		case "https", "grpcs", "tls":
		default:
			return "", "", fmt.Errorf("admin rpc endpoint must use tls")
		}
		host := strings.TrimSpace(parsed.Hostname())
		if host == "" {
			return "", "", fmt.Errorf("cannot resolve server name from admin rpc endpoint %q", endpoint)
		}
		port := strings.TrimSpace(parsed.Port())
		if port == "" {
			port = "443"
		}
		return net.JoinHostPort(host, port), host, nil
	}
	host, port, splitErr := net.SplitHostPort(raw)
	if splitErr != nil {
		host = strings.Trim(raw, "[]")
		port = "443"
	}
	host = strings.Trim(strings.TrimSpace(host), "[]")
	if host == "" {
		return "", "", fmt.Errorf("cannot resolve server name from admin rpc endpoint %q", endpoint)
	}
	if strings.TrimSpace(port) == "" {
		port = "443"
	}
	return net.JoinHostPort(host, port), host, nil
}

func readStructString(fields map[string]*structpb.Value, key string) string {
	if fields == nil {
		return ""
	}
	value := fields[key]
	if value == nil {
		return ""
	}
	return strings.TrimSpace(value.GetStringValue())
}

func readStructInt(fields map[string]*structpb.Value, key string) int {
	if fields == nil {
		return 0
	}
	value := fields[key]
	if value == nil {
		return 0
	}
	return int(value.GetNumberValue())
}
