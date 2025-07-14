package main

import (
	"context"
	"log/slog"
	"os"

	"github.com/m-mizutani/ctxlog"
)

func main() {
	// Create context and logger
	ctx := context.Background()
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}))

	// Embed logger in context
	ctx = ctxlog.With(ctx, logger)

	// Call function that extracts logger from context
	processRequest(ctx, "user123")
}

func processRequest(ctx context.Context, userID string) {
	// Extract logger from context
	logger := ctxlog.From(ctx)

	// Add context-specific fields
	logger = logger.With("user_id", userID)

	logger.Info("Processing request")
	
	// Pass context to other functions
	validateUser(ctx, userID)
	handleRequest(ctx, userID)
	
	logger.Info("Request completed")
}

func validateUser(ctx context.Context, userID string) {
	logger := ctxlog.From(ctx)
	logger.Debug("Validating user", "user_id", userID)
	// User validation logic...
}

func handleRequest(ctx context.Context, userID string) {
	logger := ctxlog.From(ctx)
	logger.Info("Handling request", "user_id", userID)
	// Request handling logic...
}