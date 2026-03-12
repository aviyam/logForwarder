package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/logforwarder/logforwarder/internal/output"
	"github.com/logforwarder/logforwarder/internal/server"
)

func main() {
	// Configure internal logger to write to stderr
	logger := slog.New(slog.NewJSONHandler(os.Stderr, nil))
	slog.SetDefault(logger)

	lumberjackAddr := getEnv("LUMBERJACK_ADDR", ":5044")
	fluentAddr := getEnv("FLUENT_ADDR", ":24224")

	handler := output.NewHandler()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	lumberjackSrv := server.NewLumberjackServer(lumberjackAddr, handler)
	if err := lumberjackSrv.Start(ctx); err != nil {
		slog.Error("failed to start lumberjack server", "error", err)
		os.Exit(1)
	}

	fluentSrv := server.NewFluentServer(fluentAddr, handler)
	if err := fluentSrv.Start(ctx); err != nil {
		slog.Error("failed to start fluent server", "error", err)
		os.Exit(1)
	}

	// Wait for termination signal
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	sig := <-sigChan
	slog.Info("shutting down", "signal", sig)
	cancel()

	// Graceful shutdown
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer shutdownCancel()

	_ = lumberjackSrv.Close()
	_ = fluentSrv.Close()

	slog.Info("shutdown complete")
	<-shutdownCtx.Done()
}

func getEnv(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}
	return fallback
}
