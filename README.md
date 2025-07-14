# ctxlog

[![GoDoc](https://godoc.org/github.com/m-mizutani/ctxlog?status.svg)](https://godoc.org/github.com/m-mizutani/ctxlog)

A Go library for embedding and extracting `*slog.Logger` instances in `context.Context` with conditional activation and hierarchical scoping.

## Overview

Large Go applications face two critical logging challenges:

**Context Preservation**: As your application grows, maintaining logging context across function boundaries becomes essential. Without proper logger propagation, you lose valuable correlation between related operations, making debugging and monitoring difficult.

**Noise Reduction**: In complex codebases, enabling all logging creates overwhelming noise. You need granular control to focus on specific components or features without modifying code or redeploying.

`ctxlog` addresses these challenges by:
- **Seamless Context Propagation**: Embed loggers with request IDs, trace IDs, or user context in `context.Context` and automatically propagate them through your entire call stack
- **Surgical Debugging**: Use scopes to enable detailed logging only for specific components, features, or code paths without affecting the rest of your application
- **Zero-Code Activation**: Control logging granularity through environment variables or runtime configuration without touching your application code

## Features

### Core Functionality
- **Context-based logger propagation**: Embed and extract loggers from context
- **Conditional activation**: Enable/disable logging based on environment variables or log levels
- **Hierarchical scoping**: Create parent-child scope relationships with inheritance

### Advanced Control
- **Probabilistic sampling**: Reduce log volume with configurable sampling rates
  - Crypto-secure random (default) or fast pseudo-random for performance
- **Dynamic control**: Runtime scope activation/deactivation
- **Test utilities**: Capture log output for testing

## Installation

```bash
go get github.com/m-mizutani/ctxlog
```

## Basic Usage

### Context Logger Propagation

```go
ctx := context.Background()
logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))

// Embed logger in context
ctx = ctxlog.With(ctx, logger)

// Extract logger from context
logger = ctxlog.From(ctx)
logger.Info("Hello, World!")
```

### Scope-based Conditional Logging

```go
// Create scope activated by environment variable
scope := ctxlog.NewScope("api", ctxlog.EnabledBy("DEBUG_API"))

// Logger is active only when DEBUG_API is set
logger := ctxlog.From(ctx, scope)
logger.Info("API call started") // Only logged if DEBUG_API exists
```

### Log Level Activation

```go
// Create scope activated when log level >= Warn
scope := ctxlog.NewScope("errors", ctxlog.EnabledMinLevel(slog.LevelWarn))

// Set current log level in context
ctx = ctxlog.WithLogLevel(ctx, slog.LevelError)

// Logger is active because Error >= Warn
logger := ctxlog.From(ctx, scope)
logger.Info("This will be logged")
```

### Multiple Activation Conditions

```go
// Scope is active if ANY condition is met
scope := ctxlog.NewScope("debug",
    ctxlog.EnabledBy("DEBUG_MODE", "VERBOSE"),  // Environment variables
    ctxlog.EnabledMinLevel(slog.LevelWarn))     // Log level threshold

// Active if DEBUG_MODE OR VERBOSE is set OR log level >= Warn
logger := ctxlog.From(ctx, scope)
```

### Hierarchical Scopes

```go
// Create parent scope
apiScope := ctxlog.NewScope("api", ctxlog.EnabledBy("DEBUG_API"))

// Create child scope - inherits parent activation
userScope := apiScope.NewChild("user", ctxlog.EnabledBy("DEBUG_USER"))

// Child is active if parent is active OR own conditions are met
logger := ctxlog.From(ctx, userScope)
```

### Dynamic Scope Control

```go
scope := ctxlog.NewScope("feature")

// Enable scope in context
ctx = ctxlog.EnableScope(ctx, scope)
logger := ctxlog.From(ctx, scope) // Active

// Enable scope globally
ctxlog.EnableScopeGlobal(scope)
logger = ctxlog.From(ctx, scope) // Active globally

// Disable scope globally
ctxlog.DisableScopeGlobal(scope)
```

### Probabilistic Sampling

```go
// Log only 10% of messages
logger := ctxlog.From(ctx, ctxlog.WithSampling(0.1))

// Use fast pseudo-random for better performance
logger = ctxlog.From(ctx, 
    ctxlog.WithSampling(0.1),
    ctxlog.WithFastRand()) // Uses math/rand instead of crypto/rand
```

### Conditional Logging

```go
// Enable logging based on custom condition
logger := ctxlog.From(ctx, ctxlog.WithCond(func() bool {
    return time.Now().Hour() < 12 // Only log in morning
}))
```

### Test Utilities

```go
func TestMyFunction(t *testing.T) {
    ctx := context.Background()
    
    // Capture log output
    ctx, capture := ctxlog.NewCapture(ctx)
    
    // Run function that logs
    MyFunction(ctx)
    
    // Verify log messages
    messages := capture.Messages()
    if len(messages) == 0 {
        t.Error("Expected log messages")
    }
}
```

## Scope Activation Logic

Scopes use OR logic for activation conditions. A scope is active if ANY of these conditions are met:

1. **Dynamic enablement**: `EnableScope(ctx, scope)` or `EnableScopeGlobal(scope)`
2. **Parent activation**: Parent scope is active (for child scopes)
3. **Environment variables**: Any specified environment variable exists (via `EnabledBy`)
4. **Log level**: Current log level >= specified minimum level (via `EnabledMinLevel`)

### Environment Variable Behavior

- Multiple environment variables are checked with OR logic
- Variable existence matters, not value (`export DEBUG=""` still activates)
- Uses `os.LookupEnv()` for checking

### Log Level Behavior

- Compares current context log level with scope minimum level
- Current level set via `ctxlog.WithLogLevel(ctx, level)`
- Does not affect which log methods you can call
- Only controls scope activation

## Performance Considerations

- **Crypto-secure random**: Default sampling uses `crypto/rand` for security
- **Fast random**: Use `WithFastRand()` with sampling for better performance
- **Buffered generation**: Crypto random numbers are buffered for efficiency
- **Scope caching**: Scope activation results are cached per context

## Examples

See the `examples/` directory for complete working examples:

- `examples/basic/` - Basic logger propagation
- `examples/scopes/` - Scope-based activation
- `examples/sampling/` - Probabilistic sampling
- `examples/hierarchical/` - Hierarchical scopes
- `examples/testing/` - Test utilities

## License

Apache License 2.0