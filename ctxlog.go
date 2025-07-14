package ctxlog

import (
	"context"
	"crypto/rand"
	"encoding/binary"
	"log/slog"
	mathrand "math/rand/v2"
	"sync"
)

type ctxLoggerKey struct{}

var loggerKey = ctxLoggerKey{} //nolint:gochecknoglobals // Required for context key

// From extracts a logger from the context with optional configuration.
// If no logger is found, returns slog.Default().
func From(ctx context.Context, options ...Option) *slog.Logger {
	cfg := &config{}
	for _, opt := range options {
		opt.apply(cfg)
	}

	baseLogger := slog.Default()
	if logger, ok := ctx.Value(loggerKey).(*slog.Logger); ok {
		baseLogger = logger
	}

	// Check scope activation
	if cfg.scope != nil {
		if !cfg.scope.isActive(ctx) {
			return createDiscardLogger()
		}
		// Add scope field to logger
		baseLogger = baseLogger.With("ctxlog.scope", cfg.scope.name)
	}

	// Check sampling
	if cfg.sampling != nil {
		var randVal float64
		if cfg.fastRand {
			randVal = fastRandFloat64()
		} else {
			randVal = cryptoRandFloat64()
		}
		if randVal > *cfg.sampling {
			return createDiscardLogger()
		}
	}

	// Check condition
	if cfg.condition != nil {
		if !cfg.condition() {
			return createDiscardLogger()
		}
	}

	return baseLogger
}

// With embeds a logger into the context and returns a new context.
func With(ctx context.Context, logger *slog.Logger) context.Context {
	return context.WithValue(ctx, loggerKey, logger)
}

// randPool provides buffered cryptographically secure random numbers.
type randPool struct {
	mu     sync.Mutex
	buffer []uint64
	pos    int
}

const (
	randPoolSize         = 256 // Buffer size for random numbers
	ieee754MantissaBits  = 53  // IEEE 754 double precision mantissa bits
	ieee754MantissaShift = 11  // Bit shift to extract mantissa (64 - 53)
)

var globalRandPool = &randPool{ //nolint:gochecknoglobals // Required for performance buffering
	buffer: make([]uint64, randPoolSize),
	pos:    randPoolSize, // Start with empty buffer to trigger initial fill
}

// refillBuffer fills the buffer with new random numbers.
func (rp *randPool) refillBuffer() error {
	buf := make([]byte, randPoolSize*8)
	if _, err := rand.Read(buf); err != nil {
		return err
	}

	for i := range randPoolSize {
		rp.buffer[i] = binary.BigEndian.Uint64(buf[i*8 : (i+1)*8])
	}
	rp.pos = 0
	return nil
}

// getUint64 returns a random uint64 from the buffer.
func (rp *randPool) getUint64() uint64 {
	rp.mu.Lock()
	defer rp.mu.Unlock()

	if rp.pos >= len(rp.buffer) {
		if err := rp.refillBuffer(); err != nil {
			// Fallback to 0 if crypto/rand fails
			return 0
		}
	}

	val := rp.buffer[rp.pos]
	rp.pos++
	return val
}

// fastRandFloat64 generates a fast pseudo-random float64 between 0 and 1.
func fastRandFloat64() float64 {
	return mathrand.Float64() // #nosec G404 - intentionally using fast pseudo-random for performance
}

// cryptoRandFloat64 generates a cryptographically secure random float64 between 0 and 1.
func cryptoRandFloat64() float64 {
	// Use bit shifting to avoid division by zero and improve performance
	// Take the upper bits for IEEE 754 double precision mantissa
	return float64(globalRandPool.getUint64()>>ieee754MantissaShift) * (1.0 / (1 << ieee754MantissaBits))
}
