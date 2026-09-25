package game

import (
	"testing"

	"github.com/google/uuid"
)

// scoped_effects_bench_test.go measures the layer-pass cost of ADR 0041
// phase 3's data records (#1558; tier 3a retired the legacy closure
// statics they were measured against): 100 effects on one creature, and one full recompute
// per iteration — what a board that has to recompute pays for them.

func benchScopedBoard(b *testing.B) (*Game, uuid.UUID) {
	b.Helper()
	g := newActiveGameWithSeats(b, 2)
	return g, pushScopedTestCreature(g, g.Seats[0].ID, 2, 2)
}

func benchRecompute(b *testing.B, g *Game) {
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		g.WithWriteLock(func() {
			g.layerVersion.Add(1)
			g.RecomputeLayersIfStaleLocked()
		})
	}
}

// BenchmarkScopedEffectRecompute100 is 100 data records on one
// creature: what the adapter costs a recompute.
func BenchmarkScopedEffectRecompute100(b *testing.B) {
	g, bear := benchScopedBoard(b)
	g.WithWriteLock(func() {
		for i := 0; i < 100; i++ {
			g.RegisterScopedEffectForEffect(uuid.Nil, g.PinnedObjectsLocked(bear),
				[]Mod{ModifyPTMod(1, 1)}, IndefiniteDuration(), "bench")
		}
	})
	benchRecompute(b, g)
}
