package ctxlog

import (
	"context"
	"os"
	"sync"
)

// Scope system provides hierarchical and conditional logger activation.
//
// COMPLETE USAGE EXAMPLE:
//
//   // 1. Create scopes with different activation conditions
//   apiScope := ctxlog.NewScope("api",
//       ctxlog.EnabledBy("DEBUG_API", "VERBOSE_API"))    // Active if either env var is set
//
//   dbScope := ctxlog.NewScope("database",
//       ctxlog.EnabledBy("DEBUG_DB"))                    // Active only if DEBUG_DB is set
//
//   errorScope := ctxlog.NewScope("errors",
//       ctxlog.EnabledBy("DEBUG_ERRORS"))               // Active only if DEBUG_ERRORS is set
//
//   // 2. Create hierarchical scopes
//   userApiScope := apiScope.NewChild("user",
//       ctxlog.EnabledBy("DEBUG_USER_API"))              // Inherits parent activation + own conditions
//
//   // 3. Set up context
//   ctx := context.Background()
//
//   // 4. Use scopes in different scenarios
//
//   // Scenario A: Environment variable activation
//   os.Setenv("DEBUG_API", "1")
//   logger := ctxlog.From(ctx, apiScope)                 // Active (DEBUG_API is set)
//   logger.Info("API call started")                      // This will be logged
//
//   // Scenario B: Multiple conditions
//   os.Unsetenv("DEBUG_API")                             // Remove env var
//   os.Setenv("VERBOSE_API", "")                         // Set different env var (empty value)
//   logger = ctxlog.From(ctx, apiScope)                  // Active (VERBOSE_API exists)
//   logger.Error("API error occurred")                   // This will be logged
//
//   // Scenario C: Hierarchical activation
//   ctx = ctxlog.EnableScope(ctx, apiScope)              // Enable parent scope
//   logger = ctxlog.From(ctx, userApiScope)              // Active (parent is active)
//   logger.Debug("User API debug")                       // This will be logged
//
//   // Scenario D: Inactive scope (all conditions fail)
//   ctx = context.Background()                           // Reset context
//   os.Unsetenv("DEBUG_API")                             // No env vars
//   os.Unsetenv("VERBOSE_API")
//   logger = ctxlog.From(ctx, apiScope)                  // Inactive (no env vars, not enabled)
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
	parent   *Scope
	children []*Scope
	mu       sync.RWMutex
}

// ScopeOption defines a functional option for Scope configuration
type ScopeOption func(*scopeConfig)

// scopeConfig holds configuration for Scope creation
type scopeConfig struct {
	envVars []string
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

// NewScope creates a new scope with the given name and options.
//
// Scope activation behavior:
// Multiple options are combined with OR logic - if ANY condition is met, the scope is active.
// Available activation conditions:
// 1. Dynamic enablement via EnableScope(ctx, scope) or EnableScopeGlobal(scope)
// 2. Parent scope activation (children inherit parent's active state)
// 3. Environment variable existence (via EnabledBy option)
//
// Combined options examples:
//
// Example 1: Multiple environment variables
//
//	scope := ctxlog.NewScope("api", ctxlog.EnabledBy("DEBUG_API", "TRACE_API", "DEV_MODE"))
//	// Active if ANY of DEBUG_API, TRACE_API, or DEV_MODE is set
//
// Example 2: Single environment variable
//
//	scope := ctxlog.NewScope("debug", ctxlog.EnabledBy("DEBUG_MODE"))
//	// Active if DEBUG_MODE env var is set (any value, even empty)
//
// Example 3: No options (manual activation only)
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
		name:    name,
		envVars: cfg.envVars,
	}

	globalScopes[name] = scope
	return scope
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

// isActive checks if the scope is active based on context, environment variables or dynamic enablement.
//
// Activation priority (checked in this order):
// 1. Context-based dynamic enablement (EnableScope)
// 2. Global dynamic enablement (EnableScopeGlobal)
// 3. Parent scope activation (recursive check)
// 4. Environment variable existence (EnabledBy option)
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
