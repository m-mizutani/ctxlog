package ctxlog_test

import (
	"context"
	"log/slog"
	"os"
	"testing"

	"github.com/m-mizutani/ctxlog"
)

func TestFrom(t *testing.T) {
	ctx := context.Background()

	// Test with no logger in context
	logger := ctxlog.From(ctx)
	if logger != slog.Default() {
		t.Error("Expected default logger when no logger in context")
	}

	// Test with logger in context
	customLogger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	ctxWithLogger := ctxlog.With(ctx, customLogger)
	logger = ctxlog.From(ctxWithLogger)
	if logger != customLogger {
		t.Error("Expected custom logger from context")
	}
}

func TestWith(t *testing.T) {
	ctx := context.Background()
	customLogger := slog.New(slog.NewTextHandler(os.Stdout, nil))

	newCtx := ctxlog.With(ctx, customLogger)
	retrievedLogger := ctxlog.From(newCtx)

	if retrievedLogger != customLogger {
		t.Error("Logger not properly stored in context")
	}
}

func TestMultipleOptions(t *testing.T) {
	// Test combining multiple options
	scope := ctxlog.NewScope("test-multi", ctxlog.EnabledBy("TEST_MULTI"))
	ctx := context.Background()
	ctx = ctxlog.EnableScope(ctx, scope)

	// Should work with active scope and true condition
	logger := ctxlog.From(ctx, scope, ctxlog.WithCond(func() bool { return true }))
	if !logger.Enabled(ctx, slog.LevelInfo) {
		t.Error("Multiple options should work together when all conditions are met")
	}

	// Should fail with active scope but false condition
	logger = ctxlog.From(ctx, scope, ctxlog.WithCond(func() bool { return false }))
	if logger.Enabled(ctx, slog.LevelInfo) {
		t.Error("Multiple options should fail when any condition is not met")
	}
}
