package ctxlog_test

import (
	"testing"

	"github.com/m-mizutani/ctxlog"
)

func BenchmarkSampling(b *testing.B) {
	ctx := b.Context()
	scope := ctxlog.NewScope("bench", ctxlog.EnabledBy("BENCH"))
	ctx = ctxlog.EnableScope(ctx, scope)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		logger := ctxlog.From(ctx, scope, ctxlog.WithSampling(0.5))
		logger.Info("benchmark message")
	}
}

func BenchmarkSamplingParallel(b *testing.B) {
	ctx := b.Context()
	scope := ctxlog.NewScope("bench-parallel", ctxlog.EnabledBy("BENCH_PARALLEL"))
	ctx = ctxlog.EnableScope(ctx, scope)

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			logger := ctxlog.From(ctx, scope, ctxlog.WithSampling(0.5))
			logger.Info("benchmark message")
		}
	})
}

func BenchmarkWithoutSampling(b *testing.B) {
	ctx := b.Context()
	scope := ctxlog.NewScope("bench-no-sampling", ctxlog.EnabledBy("BENCH_NO_SAMPLING"))
	ctx = ctxlog.EnableScope(ctx, scope)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		logger := ctxlog.From(ctx, scope)
		logger.Info("benchmark message")
	}
}
