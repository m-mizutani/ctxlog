package ctxlog_test

import (
	"context"
	"log/slog"
	"testing"

	"github.com/m-mizutani/ctxlog"
)

func TestScope(t *testing.T) {
	// Test scope creation
	scope := ctxlog.NewScope("test", ctxlog.EnabledBy("TEST_ENV"))
	if scope.Name() != "test" {
		t.Error("Scope name not set correctly")
	}

	// Test scope registry
	scope2 := ctxlog.NewScope("test", ctxlog.EnabledBy("TEST_ENV2"))
	if scope != scope2 {
		t.Error("Scope registry not working, should return same instance")
	}

	// Test hierarchical scope
	childScope := scope.NewChild("child", ctxlog.EnabledBy("CHILD_ENV"))
	if childScope.Name() != "test.child" {
		t.Error("Child scope name not constructed correctly")
	}
}

func TestScopeActivation(t *testing.T) {
	scope := ctxlog.NewScope("test-activation", ctxlog.EnabledBy("TEST_ACTIVATION"))
	ctx := context.Background()

	// Test inactive scope
	logger := ctxlog.From(ctx, scope)
	if logger.Enabled(ctx, slog.LevelInfo) {
		t.Error("Scope should be inactive without environment variable")
	}

	// Test dynamic activation
	ctx = ctxlog.EnableScope(ctx, scope)

	logger = ctxlog.From(ctx, scope)
	if !logger.Enabled(ctx, slog.LevelInfo) {
		t.Error("Scope should be active when dynamically enabled")
	}
}

func TestScopeEnvironmentActivation(t *testing.T) {
	scope := ctxlog.NewScope("test-env", ctxlog.EnabledBy("TEST_ENV_ACTIVATION"))
	ctx := context.Background()

	// Set environment variable
	t.Setenv("TEST_ENV_ACTIVATION", "1")

	logger := ctxlog.From(ctx, scope)
	if !logger.Enabled(ctx, slog.LevelInfo) {
		t.Error("Scope should be active when environment variable is set")
	}
}

func TestScopeEnvironmentActivationWithEmptyValue(t *testing.T) {
	scope := ctxlog.NewScope("test-env-empty", ctxlog.EnabledBy("TEST_ENV_EMPTY"))
	ctx := context.Background()

	// Set environment variable to empty string
	t.Setenv("TEST_ENV_EMPTY", "")

	logger := ctxlog.From(ctx, scope)
	if !logger.Enabled(ctx, slog.LevelInfo) {
		t.Error("Scope should be active even when environment variable is set to empty string")
	}
}

func TestChildScopeActivation(t *testing.T) {
	// Create deeply nested hierarchy
	parentScope := ctxlog.NewScope("parent", ctxlog.EnabledBy("PARENT_ENV"))
	childScope := parentScope.NewChild("child", ctxlog.EnabledBy("CHILD_ENV"))
	grandChildScope := childScope.NewChild("grandchild", ctxlog.EnabledBy("GRANDCHILD_ENV"))
	greatGrandChildScope := grandChildScope.NewChild("greatgrandchild", ctxlog.EnabledBy("GREATGRANDCHILD_ENV"))

	ctx := context.Background()

	// Test that all levels are inactive when parent is inactive
	logger := ctxlog.From(ctx, childScope)
	if logger.Enabled(ctx, slog.LevelInfo) {
		t.Error("Child scope should be inactive when parent is inactive")
	}

	logger = ctxlog.From(ctx, grandChildScope)
	if logger.Enabled(ctx, slog.LevelInfo) {
		t.Error("Grandchild scope should be inactive when parent is inactive")
	}

	logger = ctxlog.From(ctx, greatGrandChildScope)
	if logger.Enabled(ctx, slog.LevelInfo) {
		t.Error("Great-grandchild scope should be inactive when parent is inactive")
	}

	// Activate parent scope - all descendants should automatically become active
	ctx = ctxlog.EnableScope(ctx, parentScope)

	logger = ctxlog.From(ctx, childScope)
	if !logger.Enabled(ctx, slog.LevelInfo) {
		t.Error("Child scope should be active when parent is activated")
	}

	logger = ctxlog.From(ctx, grandChildScope)
	if !logger.Enabled(ctx, slog.LevelInfo) {
		t.Error("Grandchild scope should be active when parent is activated")
	}

	logger = ctxlog.From(ctx, greatGrandChildScope)
	if !logger.Enabled(ctx, slog.LevelInfo) {
		t.Error("Great-grandchild scope should be active when parent is activated")
	}
}

func TestChildScopeEnvironmentActivation(t *testing.T) {
	// Create parent and child scopes
	parentScope := ctxlog.NewScope("env-parent", ctxlog.EnabledBy("ENV_PARENT"))
	childScope := parentScope.NewChild("env-child", ctxlog.EnabledBy("ENV_CHILD"))

	ctx := context.Background()

	// Test child activation through environment variable
	t.Setenv("ENV_CHILD", "1")

	logger := ctxlog.From(ctx, childScope)
	if !logger.Enabled(ctx, slog.LevelInfo) {
		t.Error("Child scope should be active when its environment variable is set")
	}

	// Parent should still be inactive
	logger = ctxlog.From(ctx, parentScope)
	if logger.Enabled(ctx, slog.LevelInfo) {
		t.Error("Parent scope should remain inactive when only child environment variable is set")
	}
}

func TestChildScopeParentEnvironmentActivation(t *testing.T) {
	// Create parent and child scopes
	parentScope := ctxlog.NewScope("env-parent2", ctxlog.EnabledBy("ENV_PARENT2"))
	childScope := parentScope.NewChild("env-child2", ctxlog.EnabledBy("ENV_CHILD2"))

	ctx := context.Background()

	// Test child activation through parent environment variable
	t.Setenv("ENV_PARENT2", "1")

	// Child should be active because parent is active
	logger := ctxlog.From(ctx, childScope)
	if !logger.Enabled(ctx, slog.LevelInfo) {
		t.Error("Child scope should be active when parent environment variable is set")
	}

	// Parent should also be active
	logger = ctxlog.From(ctx, parentScope)
	if !logger.Enabled(ctx, slog.LevelInfo) {
		t.Error("Parent scope should be active when its environment variable is set")
	}
}

func TestChildScopeIndependentActivation(t *testing.T) {
	// Create scopes
	parentScope := ctxlog.NewScope("independent-parent", ctxlog.EnabledBy("INDEPENDENT_PARENT"))
	childScope := parentScope.NewChild("independent-child", ctxlog.EnabledBy("INDEPENDENT_CHILD"))

	ctx := context.Background()

	// Activate only child scope directly
	ctx = ctxlog.EnableScope(ctx, childScope)

	// Child should be active
	logger := ctxlog.From(ctx, childScope)
	if !logger.Enabled(ctx, slog.LevelInfo) {
		t.Error("Child scope should be active when directly enabled")
	}

	// Parent should still be inactive
	logger = ctxlog.From(ctx, parentScope)
	if logger.Enabled(ctx, slog.LevelInfo) {
		t.Error("Parent scope should remain inactive when only child is enabled")
	}
}

func TestDeepHierarchyActivation(t *testing.T) {
	// Create 5-level deep hierarchy
	level1 := ctxlog.NewScope("level1", ctxlog.EnabledBy("LEVEL1"))
	level2 := level1.NewChild("level2", ctxlog.EnabledBy("LEVEL2"))
	level3 := level2.NewChild("level3", ctxlog.EnabledBy("LEVEL3"))
	level4 := level3.NewChild("level4", ctxlog.EnabledBy("LEVEL4"))
	level5 := level4.NewChild("level5", ctxlog.EnabledBy("LEVEL5"))

	ctx := context.Background()

	// Test activation from middle level (level3)
	ctx = ctxlog.EnableScope(ctx, level3)

	// Levels 1, 2 should be inactive
	logger := ctxlog.From(ctx, level1)
	if logger.Enabled(ctx, slog.LevelInfo) {
		t.Error("Level1 should be inactive when level3 is activated")
	}

	logger = ctxlog.From(ctx, level2)
	if logger.Enabled(ctx, slog.LevelInfo) {
		t.Error("Level2 should be inactive when level3 is activated")
	}

	// Level 3 should be active
	logger = ctxlog.From(ctx, level3)
	if !logger.Enabled(ctx, slog.LevelInfo) {
		t.Error("Level3 should be active when directly enabled")
	}

	// Levels 4, 5 should be active (children of activated level3)
	logger = ctxlog.From(ctx, level4)
	if !logger.Enabled(ctx, slog.LevelInfo) {
		t.Error("Level4 should be active when parent level3 is activated")
	}

	logger = ctxlog.From(ctx, level5)
	if !logger.Enabled(ctx, slog.LevelInfo) {
		t.Error("Level5 should be active when ancestor level3 is activated")
	}

	// Test activation from top level
	ctx = ctxlog.EnableScope(context.Background(), level1)

	// All levels should be active
	for i, scope := range []*ctxlog.Scope{level1, level2, level3, level4, level5} {
		logger = ctxlog.From(ctx, scope)
		if !logger.Enabled(ctx, slog.LevelInfo) {
			t.Errorf("Level%d should be active when top-level parent is activated", i+1)
		}
	}
}

func TestGlobalScopeOperations(t *testing.T) {
	scope1 := ctxlog.NewScope("global-test-1", ctxlog.EnabledBy("GLOBAL_TEST_1"))
	scope2 := ctxlog.NewScope("global-test-2", ctxlog.EnabledBy("GLOBAL_TEST_2"))
	ctx := context.Background()

	// Initially inactive
	logger := ctxlog.From(ctx, scope1)
	if logger.Enabled(ctx, slog.LevelInfo) {
		t.Error("Scope should be inactive initially")
	}

	// Enable globally by scope object
	ctxlog.EnableScopeGlobal(scope1, scope2)

	logger = ctxlog.From(ctx, scope1)
	if !logger.Enabled(ctx, slog.LevelInfo) {
		t.Error("Scope should be active after global enablement")
	}

	logger = ctxlog.From(ctx, scope2)
	if !logger.Enabled(ctx, slog.LevelInfo) {
		t.Error("Scope2 should be active after global enablement")
	}

	// Check enabled scopes
	enabledScopes := ctxlog.GetGlobalEnabledScopes()
	if len(enabledScopes) < 2 {
		t.Errorf("Expected at least 2 enabled scopes, got %d", len(enabledScopes))
	}

	// Disable globally
	ctxlog.DisableScopeGlobal(scope1)

	logger = ctxlog.From(ctx, scope1)
	if logger.Enabled(ctx, slog.LevelInfo) {
		t.Error("Scope should be inactive after global disablement")
	}

	// scope2 should still be active
	logger = ctxlog.From(ctx, scope2)
	if !logger.Enabled(ctx, slog.LevelInfo) {
		t.Error("Scope2 should remain active")
	}

	// Clean up
	ctxlog.DisableScopeGlobal(scope2)
}

func TestScopeLogLevelActivation(t *testing.T) {
	// Test scope activation based on log level
	scope := ctxlog.NewScope("test-level", ctxlog.EnabledMinLevel(slog.LevelWarn))
	ctx := context.Background()

	// Test with log level below threshold
	ctx = ctxlog.WithLogLevel(ctx, slog.LevelInfo)
	logger := ctxlog.From(ctx, scope)
	if logger.Enabled(ctx, slog.LevelInfo) {
		t.Error("Scope should be inactive when log level is below threshold")
	}

	// Test with log level at threshold
	ctx = ctxlog.WithLogLevel(ctx, slog.LevelWarn)
	logger = ctxlog.From(ctx, scope)
	if !logger.Enabled(ctx, slog.LevelInfo) {
		t.Error("Scope should be active when log level is at threshold")
	}

	// Test with log level above threshold
	ctx = ctxlog.WithLogLevel(ctx, slog.LevelError)
	logger = ctxlog.From(ctx, scope)
	if !logger.Enabled(ctx, slog.LevelInfo) {
		t.Error("Scope should be active when log level is above threshold")
	}
}

func TestScopeMultipleConditions(t *testing.T) {
	// Test scope with both environment variable and log level
	scope := ctxlog.NewScope("test-multi-cond",
		ctxlog.EnabledBy("TEST_MULTI_COND"),
		ctxlog.EnabledMinLevel(slog.LevelWarn))
	ctx := context.Background()

	// Test with environment variable but low log level
	t.Setenv("TEST_MULTI_COND", "1")
	ctx = ctxlog.WithLogLevel(ctx, slog.LevelInfo)
	logger := ctxlog.From(ctx, scope)
	if !logger.Enabled(ctx, slog.LevelInfo) {
		t.Error("Scope should be active when environment variable is set, regardless of log level")
	}

	// Test with high log level but no environment variable
	t.Setenv("TEST_MULTI_COND", "")
	ctx = ctxlog.WithLogLevel(ctx, slog.LevelError)
	logger = ctxlog.From(ctx, scope)
	if !logger.Enabled(ctx, slog.LevelInfo) {
		t.Error("Scope should be active when log level is above threshold, regardless of environment variable")
	}
}
