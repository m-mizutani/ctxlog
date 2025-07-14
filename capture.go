package ctxlog

import (
	"context"
	"log/slog"
	"os"
	"sync"
)

// Capture holds captured log records for testing.
type Capture struct {
	records []slog.Record
	mu      sync.RWMutex
}

// captureHandler implements slog.Handler to capture log records.
type captureHandler struct {
	capture *Capture
	base    slog.Handler
}

func (h *captureHandler) Enabled(ctx context.Context, level slog.Level) bool {
	return h.base.Enabled(ctx, level)
}

//nolint:gocritic // slog.Record must be passed by value per slog.Handler interface
func (h *captureHandler) Handle(ctx context.Context, record slog.Record) error {
	// Make a copy to avoid issues with record reuse
	recordCopy := record.Clone()

	h.capture.mu.Lock()
	h.capture.records = append(h.capture.records, recordCopy)
	h.capture.mu.Unlock()

	return h.base.Handle(ctx, record)
}

func (h *captureHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	return &captureHandler{
		capture: h.capture,
		base:    h.base.WithAttrs(attrs),
	}
}

func (h *captureHandler) WithGroup(name string) slog.Handler {
	return &captureHandler{
		capture: h.capture,
		base:    h.base.WithGroup(name),
	}
}

// NewCapture creates a new context with log capture capability.
func NewCapture(ctx context.Context) (context.Context, *Capture) {
	capture := &Capture{}

	// Create a capture handler that wraps a text handler that outputs to discard
	handler := &captureHandler{
		capture: capture,
		base:    slog.NewTextHandler(os.Stdout, nil),
	}

	logger := slog.New(handler)

	return With(ctx, logger), capture
}

// Messages returns all captured log messages.
func (c *Capture) Messages() []string {
	c.mu.RLock()
	defer c.mu.RUnlock()

	messages := make([]string, len(c.records))
	for i := range c.records {
		messages[i] = c.records[i].Message
	}

	return messages
}

// Records returns all captured log records.
func (c *Capture) Records() []slog.Record {
	c.mu.RLock()
	defer c.mu.RUnlock()

	records := make([]slog.Record, len(c.records))
	copy(records, c.records)
	return records
}
