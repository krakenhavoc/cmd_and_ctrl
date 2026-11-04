package game

import "testing"

// fresh_for_effect_test.go pins the guards on #4909's read-side
// recompute: LookupCardForEffect brings stale layers current only while
// a resolution is open, only for a caller holding the write lock, and
// never from inside a layer pass.

// staleWhileResolving settles the layers, opens a resolution and marks
// the layers stale, as a step of a resolution does.
func staleWhileResolving(g *Game) {
	g.RecomputeLayersIfStaleLocked()
	g.resolutionOpen = true
	g.layerVersion.Add(1)
}

func TestLookupCardForEffectRecomputesStaleLayersDuringAResolution(t *testing.T) {
	g := newActiveGame(t)
	id := pushCreatureFor(t, g, g.Seats[0], 2, 2)
	g.WithWriteLock(func() {
		staleWhileResolving(g)
		before := g.recomputeCount.Load()
		if _, ok := g.LookupCardForEffect(id); !ok {
			t.Fatal("creature not found")
		}
		if got := g.recomputeCount.Load() - before; got != 1 {
			t.Errorf("a stale read during a resolution ran %d passes, want 1", got)
		}
		// Fresh now: the next reads are the fast path.
		g.LookupCardForEffect(id)
		g.LookupCardPtrForEffect(id)
		g.BattlefieldCardsForEffect()
		if got := g.recomputeCount.Load() - before; got != 1 {
			t.Errorf("reads of fresh layers ran %d passes in all, want 1", got)
		}
	})
}

func TestLookupCardForEffectDoesNotRecomputeOutsideAResolution(t *testing.T) {
	g := newActiveGame(t)
	id := pushCreatureFor(t, g, g.Seats[0], 2, 2)
	g.WithWriteLock(func() {
		g.RecomputeLayersIfStaleLocked()
		g.resolutionOpen = false
		g.layerVersion.Add(1)
		before := g.recomputeCount.Load()
		g.LookupCardForEffect(id)
		g.BattlefieldCardsForEffect()
		if got := g.recomputeCount.Load() - before; got != 0 {
			t.Errorf("a read outside a resolution ran %d layer passes", got)
		}
	})
}

func TestLookupCardForEffectDoesNotRecomputeUnderTheReadLock(t *testing.T) {
	g := newActiveGame(t)
	id := pushCreatureFor(t, g, g.Seats[0], 2, 2)
	g.WithWriteLock(func() { staleWhileResolving(g) })
	before := g.recomputeCount.Load()
	g.mu.RLock()
	g.LookupCardForEffect(id)
	g.LookupCardPtrForEffect(id)
	g.BattlefieldCardsForEffect()
	g.mu.RUnlock()
	if got := g.recomputeCount.Load() - before; got != 0 {
		t.Errorf("a read-lock holder ran %d layer passes; a recompute is a write", got)
	}
}

func TestLookupCardForEffectDoesNotRecomputeInsideALayerPass(t *testing.T) {
	g := newActiveGame(t)
	id := pushCreatureFor(t, g, g.Seats[0], 2, 2)
	g.WithWriteLock(func() {
		staleWhileResolving(g)
		g.layerPassing = true
		before := g.recomputeCount.Load()
		g.LookupCardForEffect(id)
		if got := g.recomputeCount.Load() - before; got != 0 {
			t.Errorf("a read inside a pass started %d more passes", got)
		}
		g.layerPassing = false
	})
}
