package adminrpc

import (
	"aurora/internal/config"
	"context"
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"net"
	"os"
	"strings"

	"google.golang.org/grpc"
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

	tlsCfg, err := buildTLSConfig(cfg)
	if err != nil {
		return nil, err
	}
	opts := []grpc.DialOption{
		grpc.WithTransportCredentials(credentials.NewTLS(tlsCfg)),
	}

	conn, err := grpc.DialContext(ctx, endpoint, opts...)
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

func buildTLSConfig(cfg *config.AdminRPCCfg) (*tls.Config, error) {
	serverName, err := serverNameFromEndpoint(cfg.Endpoint)
	if err != nil {
		return nil, err
	}

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

func serverNameFromEndpoint(endpoint string) (string, error) {
	raw := strings.TrimSpace(endpoint)
	if raw == "" {
		return "", fmt.Errorf("admin rpc endpoint is empty")
	}

	host := raw
	if parsedHost, _, splitErr := net.SplitHostPort(raw); splitErr == nil {
		host = parsedHost
	}
	host = strings.Trim(strings.TrimSpace(host), "[]")
	if host == "" {
		return "", fmt.Errorf("cannot resolve server name from admin rpc endpoint %q", endpoint)
	}
	return host, nil
}
