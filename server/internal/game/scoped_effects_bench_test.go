package game

import (
	"testing"

	"github.com/google/uuid"
)

// scoped_effects_bench_test.go measures the layer-pass cost of ADR 0041
// phase 3's data records against the legacy closure statics they
// replace (#1558): 100 effects on one creature, and one full recompute
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

// BenchmarkScopedStaticRecompute100 is the legacy closure static the
// records replace, 100 of them on the same creature — the baseline.
func BenchmarkScopedStaticRecompute100(b *testing.B) {
	g, bear := benchScopedBoard(b)
	g.WithWriteLock(func() {
		for i := 0; i < 100; i++ {
			g.RegisterScopedStaticForEffect(StaticAbility{
				Layer:    Layer7PT,
				SubLayer: SubLayer7C_Modify,
				AppliesTo: func(target *Card, _ *Game, _ *Card) bool {
					return target.InstanceID == bear
				},
				Apply: func(c *Characteristic, _ *Card, _ *Game, _ *Card) {
					c.Power++
					c.Toughness++
				},
			}, uuid.Nil, "bench", IndefiniteDuration())
		}
	})
	benchRecompute(b, g)
}
