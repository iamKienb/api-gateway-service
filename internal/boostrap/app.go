package bootstrap

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"shopify-api-gateway/internal/boostrap/config"
	"shopify-api-gateway/internal/boostrap/module"
	"strconv"
	"time"

	configx "github.com/iamKienb/shopify-go-platform/config"
	"golang.org/x/net/http2"
	"golang.org/x/net/http2/h2c"
)

type App struct {
	logger *slog.Logger
	server *http.Server
}

func NewApp() *App {
	return &App{
		logger: slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
			Level: slog.LevelInfo,
		})),
	}
}

func (a *App) Start(ctx context.Context) error {
	cfg, err := configx.Loader[config.ApiGatewayConfig]()
	if cfg.Server.GrpcPort == 0 {
		return fmt.Errorf("config is empty: check your .env file path")
	}

	if err != nil {
		return fmt.Errorf("load config: %w", err)
	}

	adapter := module.NewAdapterModule(a.logger, cfg)

	a.server = &http.Server{
		Addr: ":" + strconv.Itoa(cfg.Server.GrpcPort),
		Handler: h2c.NewHandler(
			adapter.Mux,
			&http2.Server{},
		),
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 30 * time.Second,
		IdleTimeout:  60 * time.Second,
	}
	a.logger.Info("starting", slog.Int("port", cfg.Server.GrpcPort))

	if err := a.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		return fmt.Errorf("server: %w", err)
	}

	return nil
}

func (a *App) Stop(ctx context.Context) error {
	a.logger.Info("shutting down")

	if a.server != nil {
		if err := a.server.Shutdown(ctx); err != nil {
			return fmt.Errorf("shutdown server: %w", err)
		}
	}

	return nil
}
