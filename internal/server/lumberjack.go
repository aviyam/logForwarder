package server

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/elastic/go-lumber/server"
	"github.com/logforwarder/logforwarder/internal/output"
)

// LumberjackServer handles logs from Filebeat.
type LumberjackServer struct {
	addr    string
	handler *output.Handler
	server  server.Server
}

// NewLumberjackServer creates a new Lumberjack server.
func NewLumberjackServer(addr string, handler *output.Handler) *LumberjackServer {
	return &LumberjackServer{
		addr:    addr,
		handler: handler,
	}
}

// Start runs the Lumberjack server.
func (s *LumberjackServer) Start(ctx context.Context) error {
	srv, err := server.ListenAndServe(s.addr, server.V2(true))
	if err != nil {
		return fmt.Errorf("failed to start lumberjack server: %w", err)
	}
	s.server = srv

	slog.Info("Lumberjack server listening", "addr", s.addr)

	go func() {
		defer s.server.Close()
		for {
			select {
			case <-ctx.Done():
				return
			case batch, ok := <-s.server.ReceiveChan():
				if !ok {
					return
				}
				slog.Debug("received lumberjack batch", "count", len(batch.Events))
				for _, event := range batch.Events {
					record, ok := event.(map[string]any)
					if !ok {
						slog.Warn("received invalid lumberjack event", "type", fmt.Sprintf("%T", event))
						continue
					}
					slog.Debug("processing lumberjack record")
					if err := s.handler.Write(output.Record(record)); err != nil {
						slog.Error("failed to write output record", "error", err)
					}
				}
				batch.ACK()
			}
		}
	}()

	return nil
}

// Close stops the Lumberjack server.
func (s *LumberjackServer) Close() error {
	if s.server != nil {
		return s.server.Close()
	}
	return nil
}
