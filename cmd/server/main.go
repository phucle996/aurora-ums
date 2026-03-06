package main

import (
	adminrpc "aurora/infra/adminrpc"
	"aurora/internal/app"
	"aurora/internal/config"
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func main() {
	cfg := config.LoadConfig()

	bootstrapCtx, cancelBootstrap := context.WithTimeout(context.Background(), cfg.AdminRPC.DialTimeout)
	if cfg.AdminRPC.DialTimeout <= 0 {
		bootstrapCtx, cancelBootstrap = context.WithTimeout(context.Background(), 5*time.Second)
	}
	runtimeValues, err := adminrpc.FetchUMSRuntimeValues(bootstrapCtx, &cfg.AdminRPC)
	if err != nil {
		cancelBootstrap()
		log.Fatalf("failed to pull runtime config from admin rpc: %v", err)
	}
	if err := cfg.ApplyRuntimeValues(runtimeValues); err != nil {
		cancelBootstrap()
		log.Fatalf("failed to apply runtime config from admin rpc: %v", err)
	}
	cancelBootstrap()

	loc, err := time.LoadLocation(cfg.App.TimeZone)
	if err != nil {
		log.Fatalf("invalid timezone: %s", cfg.App.TimeZone)
	}

	time.Local = loc

	application, err := app.NewApplication(cfg)
	if err != nil {
		log.Fatalf("failed to init application: %v", err)
	}

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		if err := application.Start(cfg); err != nil {
			log.Printf("http server stopped: %v", err)
			stop <- syscall.SIGTERM
		}
	}()

	<-stop
	log.Println("shutting down application")
	application.Stop()
}
