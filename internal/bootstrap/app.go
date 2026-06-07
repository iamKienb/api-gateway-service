package bootstrap

import (
	"api-gateway/internal/bootstrap/config"
	"api-gateway/internal/bootstrap/module"
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"strconv"
	"time"

	configx "github.com/iamKienb/go-core/config"
	"github.com/rs/cors"
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
	if err != nil {
		return fmt.Errorf("load config: %w", err)
	}
	if cfg == nil || cfg.Server.GrpcPort == 0 {
		return fmt.Errorf("config is empty: check your .env file path")
	}

	adapter, err := module.NewAdapterModule(a.logger, cfg)
	if err != nil {
		return fmt.Errorf("adapter: %w", err)
	}

	c := cors.New(cors.Options{
		AllowedOrigins:   []string{"http://localhost:3000"},
		AllowedMethods:   []string{"POST", "GET", "OPTIONS"},
		AllowedHeaders:   []string{"Content-Type", "Connect-Protocol-Version", "Authorization"},
		AllowCredentials: true,
	})
	corsHandler := c.Handler(adapter.Mux)

	a.server = &http.Server{
		Addr: ":" + strconv.Itoa(cfg.Server.GrpcPort),
		Handler: h2c.NewHandler(
			corsHandler,
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
