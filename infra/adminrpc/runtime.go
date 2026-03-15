package adminrpc

import (
	runtimev1 "github.com/phucle996/aurora-proto/runtimev1"
	"aurora/internal/config"
	"context"
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"net"
	"net/url"
	"os"
	"strconv"
	"strings"

	gogrpc "google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

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

	res, err := runtimev1.NewRuntimeServiceClient(conn).GetRuntimeBootstrap(ctx, &runtimev1.GetRuntimeBootstrapRequest{
		ModuleName: "ums",
		ConfigGroups: []string{
			"app",
			"psql",
			"redis",
			"token",
		},
	})
	if err != nil {
		return nil, fmt.Errorf("invoke admin runtime rpc failed: %w", err)
	}
	root := res.GetConfig()
	if root == nil {
		return nil, fmt.Errorf("runtime rpc response missing config")
	}
	app := root.GetApp()
	psql := root.GetPsql()
	redis := root.GetRedis()
	token := root.GetToken()
	if app == nil {
		return nil, fmt.Errorf("runtime rpc response missing config.app")
	}
	if psql == nil {
		return nil, fmt.Errorf("runtime rpc response missing config.psql")
	}
	if redis == nil {
		return nil, fmt.Errorf("runtime rpc response missing config.redis")
	}
	if token == nil {
		return nil, fmt.Errorf("runtime rpc response missing config.token")
	}

	runtimeCfg := &config.UMSRuntimeBootstrap{
		App: config.UMSRuntimeBootstrapApp{
			TimeZone: app.GetTimezone(),
			LogLevel: app.GetLogLevel(),
			Port:     int(app.GetPort()),
		},
		PSQL: config.UMSRuntimeBootstrapPSQL{
			URL:     psql.GetUrl(),
			SSLMode: psql.GetSslMode(),
			Schema:  psql.GetSchema(),
		},
		Redis: config.UMSRuntimeBootstrapRedis{
			Addr:               redis.GetAddr(),
			Username:           redis.GetUsername(),
			Password:           redis.GetPassword(),
			DB:                 redis.GetDb(),
			UseTLS:             strconv.FormatBool(redis.GetUseTls()),
			CA:                 redis.GetCa(),
			ClientKey:          redis.GetClientKey(),
			ClientCert:         redis.GetClientCert(),
			InsecureSkipVerify: strconv.FormatBool(redis.GetInsecureSkipVerify()),
		},
		Token: config.UMSRuntimeBootstrapToken{
			AccessTTL:  token.GetAccessTtl(),
			RefreshTTL: token.GetRefreshTtl(),
			DeviceTTL:  token.GetDeviceTtl(),
			OttTTL:     token.GetOttTtl(),
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
			return "", "", fmt.Errorf("admin rpc endpoint must include an explicit gRPC port")
		}
		return net.JoinHostPort(host, port), host, nil
	}
	host, port, splitErr := net.SplitHostPort(raw)
	if splitErr != nil {
		return "", "", fmt.Errorf("admin rpc endpoint must be host:port")
	}
	host = strings.Trim(strings.TrimSpace(host), "[]")
	if host == "" {
		return "", "", fmt.Errorf("cannot resolve server name from admin rpc endpoint %q", endpoint)
	}
	if strings.TrimSpace(port) == "" {
		return "", "", fmt.Errorf("admin rpc endpoint must include an explicit gRPC port")
	}
	return net.JoinHostPort(host, port), host, nil
}
