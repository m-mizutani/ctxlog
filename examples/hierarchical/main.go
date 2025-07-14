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

	// Create hierarchical scopes
	appScope := ctxlog.NewScope("app", ctxlog.EnabledBy("DEBUG_APP"))
	apiScope := appScope.NewChild("api", ctxlog.EnabledBy("DEBUG_API"))
	userScope := apiScope.NewChild("user", ctxlog.EnabledBy("DEBUG_USER"))
	authScope := userScope.NewChild("auth", ctxlog.EnabledBy("DEBUG_AUTH"))

	// Example 1: Child inherits parent activation
	println("=== Example 1: Child Inherits Parent Activation ===")
	os.Setenv("DEBUG_APP", "1")

	appLogger := ctxlog.From(ctx, appScope)
	appLogger.Info("App scope message") // Active (DEBUG_APP is set)

	apiLogger := ctxlog.From(ctx, apiScope)
	apiLogger.Info("API scope message") // Active (parent is active)

	userLogger := ctxlog.From(ctx, userScope)
	userLogger.Info("User scope message") // Active (ancestor is active)

	authLogger := ctxlog.From(ctx, authScope)
	authLogger.Info("Auth scope message") // Active (ancestor is active)

	// Example 2: Child-specific activation
	println("\n=== Example 2: Child-specific Activation ===")
	os.Unsetenv("DEBUG_APP")
	os.Setenv("DEBUG_USER", "1")

	appLogger = ctxlog.From(ctx, appScope)
	appLogger.Info("App scope message") // Inactive (DEBUG_APP not set)

	apiLogger = ctxlog.From(ctx, apiScope)
	apiLogger.Info("API scope message") // Inactive (parent inactive, DEBUG_API not set)

	userLogger = ctxlog.From(ctx, userScope)
	userLogger.Info("User scope message") // Active (DEBUG_USER is set)

	authLogger = ctxlog.From(ctx, authScope)
	authLogger.Info("Auth scope message") // Active (parent is active)

	// Example 3: Mixed activation conditions
	println("\n=== Example 3: Mixed Activation Conditions ===")
	mixedScope := ctxlog.NewScope("mixed",
		ctxlog.EnabledBy("DEBUG_MIXED"),
		ctxlog.EnabledMinLevel(slog.LevelWarn))

	childMixedScope := mixedScope.NewChild("child",
		ctxlog.EnabledBy("DEBUG_CHILD"))

	// Set log level
	ctx = ctxlog.WithLogLevel(ctx, slog.LevelError)

	os.Unsetenv("DEBUG_MIXED")
	os.Unsetenv("DEBUG_CHILD")

	mixedLogger := ctxlog.From(ctx, mixedScope)
	mixedLogger.Info("Mixed scope message") // Active (Error >= Warn)

	childLogger := ctxlog.From(ctx, childMixedScope)
	childLogger.Info("Child mixed scope message") // Active (parent is active)

	// Example 4: Dynamic activation with hierarchy
	println("\n=== Example 4: Dynamic Activation with Hierarchy ===")
	os.Unsetenv("DEBUG_USER")
	ctx = ctxlog.WithLogLevel(ctx, slog.LevelDebug)

	// Enable middle-level scope
	ctx = ctxlog.EnableScope(ctx, apiScope)

	appLogger = ctxlog.From(ctx, appScope)
	appLogger.Info("App scope message") // Inactive (not enabled)

	apiLogger = ctxlog.From(ctx, apiScope)
	apiLogger.Info("API scope message") // Active (dynamically enabled)

	userLogger = ctxlog.From(ctx, userScope)
	userLogger.Info("User scope message") // Active (parent is active)

	authLogger = ctxlog.From(ctx, authScope)
	authLogger.Info("Auth scope message") // Active (ancestor is active)

	// Example 5: Deep hierarchy with multiple conditions
	println("\n=== Example 5: Deep Hierarchy with Multiple Conditions ===")
	level1 := ctxlog.NewScope("level1", ctxlog.EnabledBy("L1"))
	level2 := level1.NewChild("level2", ctxlog.EnabledBy("L2"))
	level3 := level2.NewChild("level3", ctxlog.EnabledBy("L3"))
	level4 := level3.NewChild("level4", ctxlog.EnabledBy("L4"))
	level5 := level4.NewChild("level5", ctxlog.EnabledBy("L5"))

	// Enable level 3 - all children should be active
	os.Setenv("L3", "1")

	for i, scope := range []*ctxlog.Scope{level1, level2, level3, level4, level5} {
		logger := ctxlog.From(ctx, scope)
		logger.Info("Deep hierarchy message", "level", i+1)
	}

	println("Hierarchy examples completed")
}