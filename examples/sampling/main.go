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

	// Example 1: Basic sampling
	println("=== Example 1: Basic Sampling (50%) ===")
	for i := 0; i < 10; i++ {
		logger := ctxlog.From(ctx, ctxlog.WithSampling(0.5))
		logger.Info("Sampled message", "iteration", i)
	}

	// Example 2: High-frequency sampling with fast random
	println("\n=== Example 2: High-frequency Sampling with Fast Random ===")
	for i := 0; i < 10; i++ {
		logger := ctxlog.From(ctx, 
			ctxlog.WithSampling(0.3),
			ctxlog.WithFastRand()) // Use math/rand for better performance
		logger.Info("Fast sampled message", "iteration", i)
	}

	// Example 3: Sampling with scopes
	println("\n=== Example 3: Sampling with Scopes ===")
	scope := ctxlog.NewScope("debug", ctxlog.EnabledBy("DEBUG_MODE"))
	os.Setenv("DEBUG_MODE", "1")

	for i := 0; i < 10; i++ {
		logger := ctxlog.From(ctx, scope, ctxlog.WithSampling(0.4))
		logger.Info("Scoped sampled message", "iteration", i)
	}

	// Example 4: Conditional sampling
	println("\n=== Example 4: Conditional Sampling ===")
	for i := 0; i < 10; i++ {
		logger := ctxlog.From(ctx, 
			ctxlog.WithSampling(0.6),
			ctxlog.WithCond(func() bool {
				return i%2 == 0 // Only log even iterations
			}))
		logger.Info("Conditional sampled message", "iteration", i)
	}

	// Example 5: Performance comparison
	println("\n=== Example 5: Performance Comparison ===")
	
	// Crypto-secure random (default)
	for i := 0; i < 1000; i++ {
		logger := ctxlog.From(ctx, ctxlog.WithSampling(0.01))
		logger.Info("Crypto random", "iteration", i)
	}

	// Fast pseudo-random
	for i := 0; i < 1000; i++ {
		logger := ctxlog.From(ctx, 
			ctxlog.WithSampling(0.01),
			ctxlog.WithFastRand())
		logger.Info("Fast random", "iteration", i)
	}

	println("Performance test completed")
}