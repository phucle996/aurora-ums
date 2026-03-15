package adminrpc

import (
	"aurora/internal/config"
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"fmt"
	runtimev1 "github.com/phucle996/aurora-proto/runtimev1"
	"net/url"
	"os"
	"path/filepath"
	"strings"

	gogrpc "google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

const moduleBootstrapRuntimeRole = "module-runtime"

func EnsureUMSAdminRPCClientCertificate(ctx context.Context, cfg *config.AdminRPCCfg) error {
	if cfg == nil {
		return fmt.Errorf("admin rpc config is nil")
	}
	if strings.TrimSpace(cfg.BootstrapToken) == "" {
		if certAndKeyExist(cfg.ClientCert, cfg.ClientKey) {
			return nil
		}
		return fmt.Errorf("missing admin rpc client cert/key and bootstrap token")
	}

	target, serverName, err := normalizeAdminRPCTarget(strings.TrimSpace(cfg.Endpoint))
	if err != nil {
		return err
	}
	keyPEM, csrPEM, err := generateUMSAdminRPCKeyAndCSR()
	if err != nil {
		return err
	}

	tlsCfg, err := buildBootstrapTLSConfig(cfg, serverName)
	if err != nil {
		return err
	}
	conn, err := gogrpc.NewClient(target, gogrpc.WithTransportCredentials(credentials.NewTLS(tlsCfg)))
	if err != nil {
		return fmt.Errorf("dial admin bootstrap rpc failed: %w", err)
	}
	defer conn.Close()

	resp, err := runtimev1.NewRuntimeServiceClient(conn).BootstrapModuleClient(ctx, &runtimev1.BootstrapModuleClientRequest{
		ModuleName:     "ums",
		BootstrapToken: strings.TrimSpace(cfg.BootstrapToken),
		CsrPem:         strings.TrimSpace(string(csrPEM)),
	})
	if err != nil {
		return fmt.Errorf("bootstrap module client rpc failed: %w", err)
	}

	clientCertPEM := strings.TrimSpace(resp.GetClientCertPem())
	adminServerCAPEM := strings.TrimSpace(resp.GetAdminServerCaPem())
	if clientCertPEM == "" || adminServerCAPEM == "" {
		return fmt.Errorf("bootstrap rpc response missing certificates")
	}

	if err := writeSecurePEM(cfg.CAPath, adminServerCAPEM); err != nil {
		return fmt.Errorf("write admin ca cert failed: %w", err)
	}
	if err := writeSecurePEM(cfg.ClientCert, clientCertPEM); err != nil {
		return fmt.Errorf("write admin rpc client cert failed: %w", err)
	}
	if err := writeSecurePEM(cfg.ClientKey, strings.TrimSpace(string(keyPEM))); err != nil {
		return fmt.Errorf("write admin rpc client key failed: %w", err)
	}
	return nil
}

func buildBootstrapTLSConfig(cfg *config.AdminRPCCfg, serverName string) (*tls.Config, error) {
	caPEM, err := readBootstrapCAPEM(cfg)
	if err != nil {
		return nil, err
	}
	pool := x509.NewCertPool()
	if ok := pool.AppendCertsFromPEM(caPEM); !ok {
		return nil, fmt.Errorf("invalid admin rpc ca")
	}
	return &tls.Config{
		MinVersion: tls.VersionTLS12,
		RootCAs:    pool,
		ServerName: serverName,
	}, nil
}

func readBootstrapCAPEM(cfg *config.AdminRPCCfg) ([]byte, error) {
	if cfg == nil {
		return nil, fmt.Errorf("admin rpc config is nil")
	}
	caPath := strings.TrimSpace(cfg.CAPath)
	if caPath == "" {
		return nil, fmt.Errorf("admin rpc ca path is required")
	}
	pemBytes, err := os.ReadFile(caPath)
	if err != nil {
		return nil, fmt.Errorf("read admin rpc ca failed: %w", err)
	}
	return pemBytes, nil
}

func certAndKeyExist(certPath string, keyPath string) bool {
	if strings.TrimSpace(certPath) == "" || strings.TrimSpace(keyPath) == "" {
		return false
	}
	if _, err := os.Stat(strings.TrimSpace(certPath)); err != nil {
		return false
	}
	if _, err := os.Stat(strings.TrimSpace(keyPath)); err != nil {
		return false
	}
	return true
}

func generateUMSAdminRPCKeyAndCSR() ([]byte, []byte, error) {
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return nil, nil, fmt.Errorf("generate ums admin rpc private key failed: %w", err)
	}

	uris := buildUMSIdentityURIs()
	csrDER, err := x509.CreateCertificateRequest(rand.Reader, &x509.CertificateRequest{
		Subject: pkix.Name{
			CommonName: "module:ums",
		},
		DNSNames: []string{"ums"},
		URIs:     uris,
	}, privateKey)
	if err != nil {
		return nil, nil, fmt.Errorf("create ums admin rpc csr failed: %w", err)
	}

	keyPEM := pem.EncodeToMemory(&pem.Block{Type: "RSA PRIVATE KEY", Bytes: x509.MarshalPKCS1PrivateKey(privateKey)})
	csrPEM := pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE REQUEST", Bytes: csrDER})
	if len(keyPEM) == 0 || len(csrPEM) == 0 {
		return nil, nil, fmt.Errorf("encode ums admin rpc csr/key pem failed")
	}
	return keyPEM, csrPEM, nil
}

func buildUMSIdentityURIs() []*url.URL {
	values := []string{
		"spiffe://aurora.local/module/ums",
		"spiffe://aurora.local/service/ums",
		"spiffe://aurora.local/role/" + moduleBootstrapRuntimeRole,
	}
	out := make([]*url.URL, 0, len(values))
	for _, raw := range values {
		parsed, err := url.Parse(raw)
		if err != nil {
			continue
		}
		out = append(out, parsed)
	}
	return out
}

func writeSecurePEM(path string, content string) error {
	cleanPath := strings.TrimSpace(path)
	if cleanPath == "" {
		return fmt.Errorf("tls path is empty")
	}
	if err := os.MkdirAll(filepath.Dir(cleanPath), 0o750); err != nil {
		return fmt.Errorf("create tls dir failed (%s): %w", filepath.Dir(cleanPath), err)
	}
	return os.WriteFile(cleanPath, []byte(strings.TrimSpace(content)+"\n"), 0o600)
}
