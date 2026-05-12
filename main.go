package main

import (
	bootstrap "api-gateway-module/internal/bootstrap"
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}))

	ctx, stop := signal.NotifyContext(
		context.Background(),
		syscall.SIGINT,
		syscall.SIGTERM,
	)
	defer stop()

	app := bootstrap.NewApp()

	errCh := make(chan error, 1)
	go func() { errCh <- app.Start(ctx) }()

	select {
	case <-ctx.Done():
		logger.Info("signal received")
	case err := <-errCh:
		if err != nil {
			logger.Error("failed to start", slog.Any("err", err))
			os.Exit(1)
		}
	}

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := app.Stop(shutdownCtx); err != nil {
		logger.Error("shutdown error", slog.Any("err", err))
		os.Exit(1)
	}

	logger.Info("stopped")
}
