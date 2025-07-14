package ctxlog_test

import (
	"context"
	"log/slog"
	"testing"

	"github.com/m-mizutani/ctxlog"
)

func TestSampling(t *testing.T) {
	ctx := context.Background()

	// Test sampling with rate 0 (should always discard)
	logger := ctxlog.From(ctx, ctxlog.WithSampling(0.0))
	if logger.Enabled(ctx, slog.LevelInfo) {
		t.Error("Sampling with rate 0 should always discard")
	}

	// Test sampling with rate 1 (should never discard)
	logger = ctxlog.From(ctx, ctxlog.WithSampling(1.0))
	if !logger.Enabled(ctx, slog.LevelInfo) {
		t.Error("Sampling with rate 1 should never discard")
	}
}

func TestConditionalLogging(t *testing.T) {
	ctx := context.Background()

	// Test condition false
	logger := ctxlog.From(ctx, ctxlog.WithCond(func() bool { return false }))
	if logger.Enabled(ctx, slog.LevelInfo) {
		t.Error("Conditional logging should discard when condition is false")
	}

	// Test condition true
	logger = ctxlog.From(ctx, ctxlog.WithCond(func() bool { return true }))
	if !logger.Enabled(ctx, slog.LevelInfo) {
		t.Error("Conditional logging should allow when condition is true")
	}
}
