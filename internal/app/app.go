package app

import (
	psqlinfra "aurora/infra/psql"
	redisinfra "aurora/infra/redis"
	"aurora/internal/config"
	"aurora/internal/transport/http/handler"
	"aurora/internal/transport/http/middleware"
	"aurora/pkg/logger"
	"context"
	"fmt"
	"net"
	"net/http"
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

	// -------------------
	// PostgreSQL
	// --------------------
	db, err := psqlinfra.NewPostgres(ctx, &cfg.Psql)
	if err != nil {
		cancel()
		return nil, err
	}

	// --------------------
	// Redis
	// --------------------
	redisClient, err := redisinfra.NewRedis(ctx, &cfg.Redis)
	if err != nil {
		cancel()
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
		cancel()
		return nil, err
	}

	if cfg.TokenSecretSync.Enabled && modules.Etcd != nil {
		syncer := newTokenSecretSync(modules.Etcd, modules.Token, cfg.TokenSecretSync.Prefix)
		bootstrapTimeout := cfg.TokenSecretSync.BootstrapTimeout
		if bootstrapTimeout <= 0 {
			bootstrapTimeout = 5 * time.Second
		}
		bootstrapCtx, bootstrapCancel := context.WithTimeout(ctx, bootstrapTimeout)
		if err := syncer.Bootstrap(bootstrapCtx); err != nil {
			logger.SysWarn("token.secret.sync", "bootstrap failed, fallback to local env secrets: %v", err)
		} else {
			logger.SysInfo("token.secret.sync", "bootstrap token secrets from etcd completed")
		}
		bootstrapCancel()

		go syncer.Run(ctx)
		logger.SysInfo("token.secret.sync", "watching token secrets prefix=%s", cfg.TokenSecretSync.Prefix)
	}

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

	a.Server = &http.Server{
		Handler: a.Router,
	}

	a.hc.MarkReady()

	logger.SysInfo("http", "starting server at %s", addr)

	if err := a.Server.Serve(ln); err != nil && err != http.ErrServerClosed {
		return err
	}
	return nil
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
		if a.Modules.Etcd != nil {
			if err := a.Modules.Etcd.Close(); err != nil {
				logger.SysError("shutdown.etcd", err, "etcd shutdown failed")
			}
		}
	}

	logger.SysInfo("shutdown", "application stopped cleanly")
}
