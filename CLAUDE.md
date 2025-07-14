# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

This is a Go library (`ctxlog`) that provides utilities for embedding and extracting `*slog.Logger` instances in Go's `context.Context`. The main purpose is to propagate loggers with specific IDs (like request IDs or trace IDs) through the application call stack.

## Core Architecture

The library is structured as a single package with the following key components:

- **Context Logger Management**: Core functions `From(ctx context.Context) *slog.Logger` and `With(ctx context.Context, *slog.Logger) context.Context`
- **Scope System**: Hierarchical scoping with environment variable-based activation
- **Sampling**: Probabilistic logging with configurable rates
- **Conditional Logging**: Function-based conditional logger activation
- **Test Capture**: Logger output capture for testing scenarios

## Development Commands

### Testing
```bash
go test ./...
```

### Building
```bash
go build .
```

### Module Management
```bash
go mod tidy
```

## Key Features (from design.md)

### Scope System
- Create scopes with `NewScope(name string, envVars... string)`
- Hierarchical scopes using `Scope.NewChild(name string, envVars... string)`
- Dynamic scope control with `EnableScope(scopes... Scope) func()`
- Environment variable-based activation
- Inactive scopes return discard loggers

### Sampling and Conditional Logging
- `WithSampling(r float64)` for probabilistic logging
- `WithCond(func() bool)` for conditional logging
- Both use discard loggers when conditions aren't met

### Test Utilities
- `ctx, capture := ctxlog.NewCapture(ctx)` for test log capture
- `capture.Message()` to retrieve captured log messages

## Important Constraints

- Scopes must be globally unique within the `ctxlog` package
- The library manages scope name registration internally
- Uses Go's standard `log/slog` package as the underlying logging implementation

## File Structure

- `ctxlog.go`: Main implementation (currently empty stub)
- `design.md`: Detailed feature specification in Japanese
- `go.mod`: Go module definition targeting Go 1.24.2
- `LICENSE`: Apache 2.0 license

## Restriction & Rules

- In principle, do not trust developers who use this library from outside
  - Do not export unnecessary methods, structs, and variables
  - Assume that exposed items will be changed. Never expose fields that would be problematic if changed
  - Use `export_test.go` for items that need to be exposed for testing purposes
- When making changes, before finishing the task, always:
  - Run `go vet ./...`, `go fmt ./...` to format the code
  - Run `golangci-lint run ./...` to check lint error
  - Run `gosec -quiet ./...` to check security issue
  - Run tests to ensure no impact on other code
- All comment and charactor literal in source code must be in English
- Test files should have `package {name}_test`. Do not use same package name
- Use named empty structure (e.g. `type ctxHogeKey struct{}` ) as private context key

## 3rd party packages

- This package should in principle not use 3rd party packages, only using Go's built-in packages
- If 3rd party packages are absolutely necessary, consult first
