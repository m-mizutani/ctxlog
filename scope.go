package ctxlog

import (
	"context"
	"log/slog"
	"os"
	"sync"
)

// Scope system provides hierarchical and conditional logger activation.
//
// COMPLETE USAGE EXAMPLE:
//
//   // 1. Create scopes with different activation conditions
//   apiScope := ctxlog.NewScope("api",
//       ctxlog.EnabledBy("DEBUG_API", "VERBOSE_API"),    // Active if either env var is set
//       ctxlog.EnabledMinLevel(slog.LevelWarn))          // OR if current log level >= Warn
//
//   dbScope := ctxlog.NewScope("database",
//       ctxlog.EnabledBy("DEBUG_DB"))                    // Active only if DEBUG_DB is set
//
//   errorScope := ctxlog.NewScope("errors",
//       ctxlog.EnabledMinLevel(slog.LevelError))         // Active only if current level >= Error
//
//   // 2. Create hierarchical scopes
//   userApiScope := apiScope.NewChild("user",
//       ctxlog.EnabledBy("DEBUG_USER_API"))              // Inherits parent activation + own conditions
//
//   // 3. Set up context with log level
//   ctx := context.Background()
//   ctx = ctxlog.WithLogLevel(ctx, slog.LevelInfo)       // Set current log level
//
//   // 4. Use scopes in different scenarios
//
//   // Scenario A: Environment variable activation
//   os.Setenv("DEBUG_API", "1")
//   logger := ctxlog.From(ctx, apiScope)                 // Active (DEBUG_API is set)
//   logger.Info("API call started")                      // This will be logged
//
//   // Scenario B: Log level activation
//   os.Unsetenv("DEBUG_API")                             // Remove env var
//   ctx = ctxlog.WithLogLevel(ctx, slog.LevelWarn)       // Raise log level
//   logger = ctxlog.From(ctx, apiScope)                  // Active (Warn >= Warn)
//   logger.Info("API call completed")                    // This will be logged
//
//   // Scenario C: Multiple conditions
//   ctx = ctxlog.WithLogLevel(ctx, slog.LevelDebug)      // Lower log level
//   os.Setenv("VERBOSE_API", "")                         // Set different env var (empty value)
//   logger = ctxlog.From(ctx, apiScope)                  // Active (VERBOSE_API exists)
//   logger.Error("API error occurred")                   // This will be logged
//
//   // Scenario D: Hierarchical activation
//   ctx = ctxlog.EnableScope(ctx, apiScope)              // Enable parent scope
//   logger = ctxlog.From(ctx, userApiScope)              // Active (parent is active)
//   logger.Debug("User API debug")                       // This will be logged
//
//   // Scenario E: Inactive scope (all conditions fail)
//   ctx = context.Background()                           // Reset context
//   ctx = ctxlog.WithLogLevel(ctx, slog.LevelDebug)      // Low log level
//   os.Unsetenv("DEBUG_API")                             // No env vars
//   os.Unsetenv("VERBOSE_API")
//   logger = ctxlog.From(ctx, apiScope)                  // Inactive (Debug < Warn, no env vars)
//   logger.Info("This won't be logged")                  // Returns discard logger
//
//   // 5. Global scope management
//   ctxlog.EnableScopeGlobal(dbScope)                    // Enable globally
//   logger = ctxlog.From(ctx, dbScope)                   // Active (globally enabled)
//   logger.Info("Database query")                        // This will be logged
//
//   ctxlog.DisableScopeGlobal(dbScope)                   // Disable globally
//   logger = ctxlog.From(ctx, dbScope)                   // Inactive (not globally enabled, no DEBUG_DB)
//   logger.Info("This won't be logged")                  // Returns discard logger

// Scope represents a logging scope with hierarchical support
type Scope struct {
	name     string
	envVars  []string
	logLevel *slog.Level
	parent   *Scope
	children []*Scope
	mu       sync.RWMutex
}

// ScopeOption defines a functional option for Scope configuration
type ScopeOption func(*scopeConfig)

// scopeConfig holds configuration for Scope creation
type scopeConfig struct {
	envVars  []string
	logLevel *slog.Level
}

var (
	globalScopes  = make(map[string]*Scope) //nolint:gochecknoglobals // Required for scope registry
	scopesMu      sync.RWMutex              //nolint:gochecknoglobals // Required for scope registry
	enabledScopes = make(map[string]*Scope) //nolint:gochecknoglobals // Required for scope management
	enabledMu     sync.RWMutex              //nolint:gochecknoglobals // Required for scope management
)

type ctxEnabledScopesKey struct{}

var enabledScopesKey = ctxEnabledScopesKey{} //nolint:gochecknoglobals // Required for context key

// EnabledBy creates a ScopeOption that enables scope activation via environment variables.
//
// Multiple environment variables behavior:
//   - If ANY of the specified environment variables is set (even to empty string),
//     the scope will be activated.
//   - Environment variables are checked with os.LookupEnv(), so existence matters, not value.
//
// Example:
//
//	scope := ctxlog.NewScope("api", ctxlog.EnabledBy("DEBUG_API", "VERBOSE_API"))
//	// Scope is active if either DEBUG_API OR VERBOSE_API is set
//
//	export DEBUG_API=1     # scope is active
//	export VERBOSE_API=""  # scope is active (empty value still counts)
//	unset DEBUG_API VERBOSE_API  # scope is inactive
func EnabledBy(envVars ...string) ScopeOption {
	return func(cfg *scopeConfig) {
		cfg.envVars = envVars
	}
}

// EnabledMinLevel creates a ScopeOption that enables scope activation at minimum specified log level.
//
// Log level comparison:
// - The scope is activated when the CURRENT log level in context >= specified level
// - Current log level is set via ctxlog.WithLogLevel(ctx, level)
// - This does NOT affect which log methods (Info, Warn, Error) you can call
// - This only controls scope activation, the actual logger's level is separate
//
// Log level hierarchy (from lowest to highest):
//
//	slog.LevelDebug (-4) < slog.LevelInfo (0) < slog.LevelWarn (4) < slog.LevelError (8)
//
// Example:
//
//	scope := ctxlog.NewScope("errors", ctxlog.EnabledMinLevel(slog.LevelWarn))
//	ctx = ctxlog.WithLogLevel(ctx, slog.LevelError)  // Current level: Error
//	logger := ctxlog.From(ctx, scope)  // Scope is active (Error >= Warn minimum)
//
//	// You can still call any log method regardless of scope activation level:
//	logger.Info("info message")    // Will be logged if scope is active
//	logger.Warn("warn message")    // Will be logged if scope is active
//	logger.Error("error message")  // Will be logged if scope is active
//
// Activation examples:
//
//	EnabledMinLevel(slog.LevelWarn) + WithLogLevel(ctx, slog.LevelDebug) = inactive (Debug < Warn minimum)
//	EnabledMinLevel(slog.LevelWarn) + WithLogLevel(ctx, slog.LevelWarn)  = active   (Warn >= Warn minimum)
//	EnabledMinLevel(slog.LevelWarn) + WithLogLevel(ctx, slog.LevelError) = active   (Error >= Warn minimum)
func EnabledMinLevel(level slog.Level) ScopeOption {
	return func(cfg *scopeConfig) {
		cfg.logLevel = &level
	}
}

// NewScope creates a new scope with the given name and options.
//
// Scope activation behavior:
// Multiple options are combined with OR logic - if ANY condition is met, the scope is active.
// Available activation conditions:
// 1. Dynamic enablement via EnableScope(ctx, scope) or EnableScopeGlobal(scope)
// 2. Parent scope activation (children inherit parent's active state)
// 3. Environment variable existence (via EnabledBy option)
// 4. Log level threshold (via EnableAbove option)
//
// Combined options examples:
//
// Example 1: Environment variables OR log level
//
//	scope := ctxlog.NewScope("debug",
//	    ctxlog.EnabledBy("DEBUG_MODE", "VERBOSE"),
//	    ctxlog.EnabledMinLevel(slog.LevelWarn))
//
//	// Scope is active in ANY of these cases:
//	// - DEBUG_MODE env var is set (any value, even empty)
//	// - VERBOSE env var is set (any value, even empty)
//	// - Current log level >= slog.LevelWarn
//	// - Scope is dynamically enabled via EnableScope/EnableScopeGlobal
//	// - Parent scope is active (for child scopes)
//
// Example 2: Multiple environment variables
//
//	scope := ctxlog.NewScope("api", ctxlog.EnabledBy("DEBUG_API", "TRACE_API", "DEV_MODE"))
//	// Active if ANY of DEBUG_API, TRACE_API, or DEV_MODE is set
//
// Example 3: Log level only
//
//	scope := ctxlog.NewScope("errors", ctxlog.EnabledMinLevel(slog.LevelError))
//	ctx = ctxlog.WithLogLevel(ctx, slog.LevelError)
//	logger := ctxlog.From(ctx, scope)  // Active because Error >= Error
//
// Example 4: No options (manual activation only)
//
//	scope := ctxlog.NewScope("manual")
//	// Only active via EnableScope(ctx, scope) or EnableScopeGlobal(scope)
func NewScope(name string, options ...ScopeOption) *Scope {
	scopesMu.Lock()
	defer scopesMu.Unlock()

	if existing, exists := globalScopes[name]; exists {
		return existing
	}

	cfg := &scopeConfig{}
	for _, opt := range options {
		opt(cfg)
	}

	scope := &Scope{
		name:     name,
		envVars:  cfg.envVars,
		logLevel: cfg.logLevel,
	}

	globalScopes[name] = scope
	return scope
}

// NewScopeWithEnvVars creates a new scope with environment variables (backward compatibility)
func NewScopeWithEnvVars(name string, envVars ...string) *Scope {
	return NewScope(name, EnabledBy(envVars...))
}

// NewChild creates a child scope with hierarchical naming
func (s *Scope) NewChild(name string, options ...ScopeOption) *Scope {
	fullName := s.name + "." + name
	child := NewScope(fullName, options...)
	child.parent = s

	s.mu.Lock()
	s.children = append(s.children, child)
	s.mu.Unlock()

	return child
}

// NewChildWithEnvVars creates a child scope with environment variables (backward compatibility)
func (s *Scope) NewChildWithEnvVars(name string, envVars ...string) *Scope {
	return s.NewChild(name, EnabledBy(envVars...))
}

// isActive checks if the scope is active based on context, environment variables, log level or dynamic enablement.
//
// Activation priority (checked in this order):
// 1. Context-based dynamic enablement (EnableScope)
// 2. Global dynamic enablement (EnableScopeGlobal)
// 3. Parent scope activation (recursive check)
// 4. Log level threshold (EnableAbove option)
// 5. Environment variable existence (EnabledBy option)
//
// Returns true if ANY condition is met (OR logic).
func (s *Scope) isActive(ctx context.Context) bool {
	// Check context-based enablement first
	if enabledScopes, ok := ctx.Value(enabledScopesKey).(map[string]bool); ok {
		if enabled, exists := enabledScopes[s.name]; exists && enabled {
			return true
		}
	}

	// Check global dynamic enablement
	enabledMu.RLock()
	if _, exists := enabledScopes[s.name]; exists {
		enabledMu.RUnlock()
		return true
	}
	enabledMu.RUnlock()

	// Check parent activation (parent activation enables all children)
	if s.parent != nil && s.parent.isActive(ctx) {
		return true
	}

	// Check log level activation
	if s.logLevel != nil {
		if currentLevel, ok := ctx.Value(logLevelKey).(slog.Level); ok {
			if currentLevel >= *s.logLevel {
				return true
			}
		}
	}

	// Check environment variables
	for _, envVar := range s.envVars {
		if _, exists := os.LookupEnv(envVar); exists {
			return true
		}
	}

	return false
}

// EnableScope returns a new context with the given scopes enabled
func EnableScope(ctx context.Context, scopes ...*Scope) context.Context {
	var contextScopes map[string]bool

	// Get existing enabled scopes from context
	if existing, ok := ctx.Value(enabledScopesKey).(map[string]bool); ok {
		// Copy existing map
		contextScopes = make(map[string]bool)
		for k, v := range existing {
			contextScopes[k] = v
		}
	} else {
		contextScopes = make(map[string]bool)
	}

	// Add new scopes
	for _, scope := range scopes {
		contextScopes[scope.name] = true
	}

	return context.WithValue(ctx, enabledScopesKey, contextScopes)
}

// EnableScopeGlobal dynamically enables the given scopes globally
func EnableScopeGlobal(scopes ...*Scope) {
	enabledMu.Lock()
	defer enabledMu.Unlock()

	for _, scope := range scopes {
		enabledScopes[scope.name] = scope
	}
}

// DisableScopeGlobal disables the given scopes globally
func DisableScopeGlobal(scopes ...*Scope) {
	enabledMu.Lock()
	defer enabledMu.Unlock()

	for _, scope := range scopes {
		delete(enabledScopes, scope.name)
	}
}

// GetGlobalEnabledScopes returns the globally enabled scopes
func GetGlobalEnabledScopes() []*Scope {
	enabledMu.RLock()
	defer enabledMu.RUnlock()

	scopes := make([]*Scope, 0, len(enabledScopes))
	for _, scope := range enabledScopes {
		scopes = append(scopes, scope)
	}
	return scopes
}

// Name returns the name of the scope
func (s *Scope) Name() string {
	return s.name
}

// apply implements the Option interface
func (s *Scope) apply(c *config) {
	c.scope = s
}
