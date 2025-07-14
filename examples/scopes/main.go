package main

import (
	"context"
	"log/slog"
	"os"

	"github.com/m-mizutani/ctxlog"
)

func main() {
	ctx := context.Background()
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}))
	ctx = ctxlog.With(ctx, logger)

	// Create scopes with different activation conditions
	apiScope := ctxlog.NewScope("api", ctxlog.EnabledBy("DEBUG_API"))
	dbScope := ctxlog.NewScope("database", ctxlog.EnabledBy("DEBUG_DB"))
	errorScope := ctxlog.NewScope("errors", ctxlog.EnabledMinLevel(slog.LevelError))

	// Set log level in context
	ctx = ctxlog.WithLogLevel(ctx, slog.LevelWarn)

	// Example 1: Environment variable activation
	println("=== Example 1: Environment Variable Activation ===")
	_ = os.Setenv("DEBUG_API", "1")
	
	apiLogger := ctxlog.From(ctx, apiScope)
	apiLogger.Info("API debug message") // This will be logged

	dbLogger := ctxlog.From(ctx, dbScope)
	dbLogger.Info("Database debug message") // This will NOT be logged (DEBUG_DB not set)

	// Example 2: Log level activation
	println("\n=== Example 2: Log Level Activation ===")
	ctx = ctxlog.WithLogLevel(ctx, slog.LevelError)
	
	errorLogger := ctxlog.From(ctx, errorScope)
	errorLogger.Info("Error scope message") // This will be logged (Error >= Error)

	// Example 3: Multiple conditions
	println("\n=== Example 3: Multiple Conditions ===")
	multiScope := ctxlog.NewScope("multi",
		ctxlog.EnabledBy("DEBUG_MULTI"),
		ctxlog.EnabledMinLevel(slog.LevelWarn))

	ctx = ctxlog.WithLogLevel(ctx, slog.LevelWarn)
	multiLogger := ctxlog.From(ctx, multiScope)
	multiLogger.Info("Multi-condition message") // This will be logged (Warn >= Warn)

	// Example 4: Dynamic activation
	println("\n=== Example 4: Dynamic Activation ===")
	_ = os.Unsetenv("DEBUG_API")
	ctx = ctxlog.WithLogLevel(ctx, slog.LevelDebug)
	
	// Enable scope dynamically
	ctx = ctxlog.EnableScope(ctx, apiScope)
	dynamicLogger := ctxlog.From(ctx, apiScope)
	dynamicLogger.Info("Dynamically enabled message") // This will be logged

	// Example 5: Global activation
	println("\n=== Example 5: Global Activation ===")
	ctxlog.EnableScopeGlobal(dbScope)
	
	globalLogger := ctxlog.From(ctx, dbScope)
	globalLogger.Info("Globally enabled message") // This will be logged

	// Clean up
	ctxlog.DisableScopeGlobal(dbScope)
}