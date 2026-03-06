package app

import (
	psqlinfra "aurora/infra/psql"
	redisinfra "aurora/infra/redis"
	"aurora/internal/config"
	"aurora/internal/transport/http/handler"
	"aurora/internal/transport/http/middleware"
	"aurora/pkg/logger"
	"context"
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"net"
	"net/http"
	"os"
	"time"

	"github.com/gin-gonic/gin"
)

type App struct {
	Ctx    context.Context
	Cancel context.CancelFunc

	Modules *Modules
	Router  *gin.Engine

	Server *http.Server
	hc     *handler.HealthHandler
}

// NewApplication initializes all dependencies but DOES NOT start the server.
func NewApplication(cfg *config.Config) (*App, error) {
	logger.InitLogger(&cfg.App)

	ctx, cancel := context.WithCancel(context.Background())
	cleanup := func() {
		cancel()
	}

	// -------------------
	// PostgreSQL
	// --------------------
	db, err := psqlinfra.NewPostgres(ctx, &cfg.Psql)
	if err != nil {
		cleanup()
		return nil, err
	}

	// --------------------
	// Redis
	// --------------------
	redisClient, err := redisinfra.NewRedis(ctx, &cfg.Redis)
	if err != nil {
		db.Close()
		cleanup()
		return nil, err
	}

	publisher := redisinfra.NewRedisStreamPublisher(redisClient)
	// --------------------
	// Modules
	// --------------------
	modules, err := NewModules(ctx, db, redisClient, publisher, cfg)
	if err != nil {
		_ = redisClient.Close()
		db.Close()
		cleanup()
		return nil, err
	}

	syncer := newTokenSecretSync(modules.Redis, modules.Token)
	bootstrapTimeout := tokenSecretBootstrapTimeout
	bootstrapCtx, bootstrapCancel := context.WithTimeout(ctx, bootstrapTimeout)
	bootstrapErr := syncer.Bootstrap(bootstrapCtx)
	bootstrapCancel()
	if bootstrapErr != nil {
		_ = redisClient.Close()
		db.Close()
		cleanup()
		return nil, fmt.Errorf("bootstrap token secrets from redis failed: %w", bootstrapErr)
	}
	if err := modules.Token.ValidateSecrets(); err != nil {
		_ = redisClient.Close()
		db.Close()
		cleanup()
		return nil, fmt.Errorf("redis token secrets are incomplete: %w", err)
	}
	logger.SysInfo("token.secret.sync", "bootstrap token secrets from redis completed")

	go syncer.Run(ctx)
	logger.SysInfo(
		"token.secret.sync",
		"watching redis token secret cache prefix=%s channel=%s",
		syncer.cacheKeyPrefix,
		syncer.invalidateChannel,
	)

	health := handler.NewHealthHandler(
		db,
		redisClient,
	)
	// --------------------
	// gin http framework
	// --------------------
	gin.SetMode(gin.ReleaseMode)
	router := gin.New()
	router.Use(
		middleware.RequestContext(),
		middleware.AccessLog(),
		middleware.CORS(&cfg.Cors),
		gin.Recovery(),
	)

	RegisterRoutes(router, modules, health)

	go func() {
		ticker := time.NewTicker(24 * time.Hour)
		defer ticker.Stop()

		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				cutoff := time.Now().Add(-90 * 24 * time.Hour)
				_ = modules.DeviceSvc.CleanupStaleDevices(ctx, cutoff)
			}
		}
	}()

	health.MarkReady()
	return &App{
		Ctx:     ctx,
		Cancel:  cancel,
		Modules: modules,
		hc:      health,
		Router:  router,
	}, nil
}

func (a *App) Start(cfg *config.Config) error {
	addr := fmt.Sprintf(":%d", cfg.App.Port)

	ln, err := net.Listen("tcp", addr)
	if err != nil {
		return err
	}

	tlsCfg, tlsErr := buildServerTLSConfig(cfg.App)
	if tlsErr != nil {
		_ = ln.Close()
		return tlsErr
	}

	a.Server = &http.Server{
		Handler:   a.Router,
		TLSConfig: tlsCfg,
	}
	ln = tls.NewListener(ln, tlsCfg)

	a.hc.MarkReady()

	logger.SysInfo("http", "starting https server at %s", addr)

	if err := a.Server.Serve(ln); err != nil && err != http.ErrServerClosed {
		return err
	}
	return nil
}

func buildServerTLSConfig(cfg config.AppCfg) (*tls.Config, error) {
	certFile := cfg.TLSCertPath
	keyFile := cfg.TLSKeyPath
	caFile := cfg.TLSCAPath
	if certFile == "" || keyFile == "" || caFile == "" {
		return nil, fmt.Errorf("tls paths are required")
	}

	cert, err := tls.LoadX509KeyPair(certFile, keyFile)
	if err != nil {
		return nil, fmt.Errorf("load ums tls cert/key failed: %w", err)
	}

	caPEM, err := os.ReadFile(caFile)
	if err != nil {
		return nil, fmt.Errorf("read ums tls ca failed: %w", err)
	}
	pool := x509.NewCertPool()
	if !pool.AppendCertsFromPEM(caPEM) {
		return nil, fmt.Errorf("invalid ums tls ca pem")
	}

	return &tls.Config{
		MinVersion:   tls.VersionTLS12,
		Certificates: []tls.Certificate{cert},
		ClientCAs:    pool,
		ClientAuth:   tls.RequireAndVerifyClientCert,
		NextProtos:   []string{"h2", "http/1.1"},
	}, nil
}

func (a *App) Stop() {
	logger.SysInfo("shutdown", "shutting down application")

	a.hc.MarkNotReady()

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if a.Server != nil {
		if err := a.Server.Shutdown(ctx); err != nil {
			logger.SysError("shutdown.http", err, "http shutdown failed")
		}
	}

	if a.Cancel != nil {
		a.Cancel()
	}

	if a.Modules != nil {
		// Close db
		if a.Modules.PostgresDB != nil {
			a.Modules.PostgresDB.Close()
		}
		//  Close Redis
		if a.Modules.Redis != nil {
			_ = a.Modules.Redis.Close()
		}
	}

	logger.SysInfo("shutdown", "application stopped cleanly")
}
