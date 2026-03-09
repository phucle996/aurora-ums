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

const (
	runtimeGetUMSBootstrapPath = "/admin.transport.runtime.v1.RuntimeService/GetUMSBootstrap"
)

func FetchUMSRuntimeValues(ctx context.Context, cfg *config.AdminRPCCfg) (map[string]string, error) {
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
	})
	if err != nil {
		return nil, err
	}
	res := &structpb.Struct{}
	if err := conn.Invoke(ctx, runtimeGetUMSBootstrapPath, req, res); err != nil {
		return nil, fmt.Errorf("invoke admin runtime rpc failed: %w", err)
	}

	valuesField, ok := res.GetFields()["values"]
	if !ok || valuesField == nil || valuesField.GetStructValue() == nil {
		return nil, fmt.Errorf("runtime rpc response missing values")
	}

	values := make(map[string]string)
	for key, value := range valuesField.GetStructValue().GetFields() {
		values[strings.TrimSpace(key)] = strings.TrimSpace(value.GetStringValue())
	}
	if len(values) == 0 {
		return nil, fmt.Errorf("runtime rpc values are empty")
	}
	return values, nil
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
