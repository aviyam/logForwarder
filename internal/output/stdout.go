package output

import (
	"encoding/json"
	"os"
	"sync"
)

// Record represents a single log record.
type Record map[string]any

// Handler manages writing records to stdout.
type Handler struct {
	mu      sync.Mutex
	encoder *json.Encoder
}

// NewHandler creates a new output handler.
func NewHandler() *Handler {
	return &Handler{
		encoder: json.NewEncoder(os.Stdout),
	}
}

// Write writes a record to stdout as JSON.
func (h *Handler) Write(record Record) error {
	h.mu.Lock()
	defer h.mu.Unlock()
	return h.encoder.Encode(record)
}
