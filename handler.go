package ctxlog

import (
	"context"
	"log/slog"
)

// discardHandler creates a handler that discards all log records.
type discardHandler struct{}

func (h *discardHandler) Enabled(context.Context, slog.Level) bool {
	return false
}

func (h *discardHandler) Handle(context.Context, slog.Record) error {
	return nil
}

func (h *discardHandler) WithAttrs(_ []slog.Attr) slog.Handler {
	return h
}

func (h *discardHandler) WithGroup(_ string) slog.Handler {
	return h
}

// createDiscardLogger creates a logger that discards all output.
func createDiscardLogger() *slog.Logger {
	return slog.New(&discardHandler{})
}
