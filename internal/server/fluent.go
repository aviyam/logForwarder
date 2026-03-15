package server

import (
	"bytes"
	"compress/gzip"
	"context"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net"

	"github.com/logforwarder/logforwarder/internal/output"
	"github.com/vmihailenco/msgpack/v5"
)

// FluentServer handles logs from Fluentd/Fluent Bit using the Forward protocol.
type FluentServer struct {
	addr     string
	handler  *output.Handler
	listener net.Listener
}

// NewFluentServer creates a new Fluent server.
func NewFluentServer(addr string, handler *output.Handler) *FluentServer {
	return &FluentServer{
		addr:    addr,
		handler: handler,
	}
}

// Start runs the Fluent server.
func (s *FluentServer) Start(ctx context.Context) error {
	l, err := net.Listen("tcp", s.addr)
	if err != nil {
		return fmt.Errorf("failed to start fluent server: %w", err)
	}
	s.listener = l

	slog.Info("Fluent Forward server listening", "addr", s.addr)

	go func() {
		for {
			conn, err := s.listener.Accept()
			if err != nil {
				if errors.Is(err, net.ErrClosed) {
					return
				}
				slog.Error("failed to accept fluent connection", "error", err)
				continue
			}
			go func(c net.Conn) {
				slog.Debug("accepted fluent connection", "remote_addr", c.RemoteAddr())
				s.handleConnection(c)
			}(conn)
		}
	}()

	return nil
}

func (s *FluentServer) handleConnection(conn net.Conn) {
	defer conn.Close()
	decoder := msgpack.NewDecoder(conn)

	for {
		var msg []any
		if err := decoder.Decode(&msg); err != nil {
			if errors.Is(err, io.EOF) {
				slog.Debug("fluent connection closed by client", "remote_addr", conn.RemoteAddr())
				return
			}
			slog.Error("failed to decode fluent message", "error", err)
			return
		}

		if len(msg) < 2 {
			slog.Warn("received invalid fluent message (too short)")
			continue
		}

		tag, ok := msg[0].(string)
		if !ok {
			slog.Warn("received invalid fluent message (tag is not string)")
			continue
		}

		slog.Debug("received fluent message", "tag", tag, "remote_addr", conn.RemoteAddr())

		switch entries := msg[1].(type) {
		case []any:
			// Forward mode: [[time, record], [time, record], ...]
			for _, e := range entries {
				entry, ok := e.([]any)
				if !ok || len(entry) < 2 {
					continue
				}
				record, ok := entry[1].(map[string]any)
				if !ok {
					continue
				}
				record["_tag"] = tag
				if err := s.handler.Write(output.Record(record)); err != nil {
					slog.Error("failed to write output record", "error", err)
				}
			}
		case map[string]any:
			// Message mode: [tag, time, record, option]
			if len(msg) >= 3 {
				record, ok := msg[2].(map[string]any)
				if ok {
					record["_tag"] = tag
					if err := s.handler.Write(output.Record(record)); err != nil {
						slog.Error("failed to write output record", "error", err)
					}
				}
			}
		case []byte:
			// PackedForward or CompressedPackedForward mode: [tag, binary, option]
			var reader io.Reader
			reader = bytes.NewReader(entries)

			// Check if compressed (if option has "compressed" : "gzip")
			if len(msg) >= 3 {
				if opt, ok := msg[2].(map[string]any); ok {
					if comp, ok := opt["compressed"].(string); ok && comp == "gzip" {
						gz, err := gzip.NewReader(reader)
						if err != nil {
							slog.Error("failed to create gzip reader", "error", err)
							continue
						}
						defer gz.Close()
						reader = gz
					}
				}
			}

			innerDecoder := msgpack.NewDecoder(reader)
			for {
				var entry []any
				if err := innerDecoder.Decode(&entry); err != nil {
					if errors.Is(err, io.EOF) {
						break
					}
					slog.Error("failed to decode packed fluent entry", "error", err)
					break
				}
				if len(entry) < 2 {
					continue
				}
				record, ok := entry[1].(map[string]any)
				if !ok {
					continue
				}
				record["_tag"] = tag
				if err := s.handler.Write(output.Record(record)); err != nil {
					slog.Error("failed to write output record", "error", err)
				}
			}
		default:
			slog.Warn("received unsupported fluent message mode", "type", fmt.Sprintf("%T", msg[1]))
		}
	}
}

// Close stops the Fluent server.
func (s *FluentServer) Close() error {
	if s.listener != nil {
		return s.listener.Close()
	}
	return nil
}
