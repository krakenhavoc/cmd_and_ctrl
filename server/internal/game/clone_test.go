package game

import (
	"testing"

	"github.com/google/uuid"
)

// clone_test.go covers undo-snapshot completeness: Clone /
// RestoreFrom must capture and roll back the per-card knowledge
// sets, the S17 replacement registries, and the S16 layer-engine
// staleness so an undo really rewinds the whole observable state.

// findCardInZone returns a pointer to the card with the given
// instance ID inside z, or nil. Test-local; production code has its
// own zone scanners.
func findCardInZone(z *Zone, id uuid.UUID) *Card {
	for i := range z.Cards {
		if z.Cards[i].InstanceID == id {
			return &z.Cards[i]
		}
	}
	return nil
}

// TestCloneIsolatesKnownBy proves a reveal that happens AFTER a
// snapshot does not leak into the snapshot (the clone must deep-copy
// the KnownBy map, not alias it), and that RestoreFrom rolls the
// knowledge back.
func TestCloneIsolatesKnownBy(t *testing.T) {
	g := newActiveGame(t)
	p0, p1 := g.Seats[0], g.Seats[1]
	if p0.Hand.Size() == 0 {
		t.Fatal("setup: p0 has no opening hand")
	}
	cardID := p0.Hand.Cards[0].InstanceID

	snap := g.Clone()

	// Reveal the hand card to the opponent on the live game
	// (Thoughtseize-style) after the snapshot was taken.
	g.WithWriteLock(func() {
		if c := findCardInZone(g.Seats[0].Hand, cardID); c != nil {
			c.AddKnower(p1.ID)
		} else {
			t.Error("live card vanished from hand")
		}
	})

	// The snapshot must not have learned about the reveal.
	snapCard := findCardInZone(snap.Seats[0].Hand, cardID)
	if snapCard == nil {
		t.Fatal("snapshot lost the hand card")
	}
	if snapCard.IsKnownTo(p1.ID) {
		t.Error("post-snapshot reveal leaked into the clone's KnownBy map")
	}
	if !snapCard.IsKnownTo(p0.ID) {
		t.Error("clone dropped the owner from KnownBy")
	}

	// Undo: restoring the snapshot must roll the reveal back.
	g.WithWriteLock(func() { g.RestoreFrom(snap) })
	liveCard := findCardInZone(g.Seats[0].Hand, cardID)
	if liveCard == nil {
		t.Fatal("restored game lost the hand card")
	}
	if liveCard.IsKnownTo(p1.ID) {
		t.Error("RestoreFrom did not roll back the reveal")
	}
	if !liveCard.IsKnownTo(p0.ID) {
		t.Error("RestoreFrom dropped the owner's knowledge")
	}
}

// TestRestoreFromRollsBackScopedReplacements is the Fog-undo repro:
// register a Fog after a snapshot, then restore — the replacement must
// be gone. (Before ADR 0041 tier 3b it lived in TurnScopedReplacements,
// which RestoreFrom once omitted entirely, so an undone Fog kept
// preventing damage.)
func TestRestoreFromRollsBackScopedReplacements(t *testing.T) {
	g := newActiveGame(t)
	snap := g.Clone()

	g.WithWriteLock(func() {
		g.PreventCombatDamageThisTurnForEffect(uuid.Nil, uuid.Nil, "Fog: prevent all combat damage this turn")
	})
	if len(g.ScopedEffects) != 1 {
		t.Fatalf("setup: the Fog registered %d records, want 1", len(g.ScopedEffects))
	}

	g.WithWriteLock(func() { g.RestoreFrom(snap) })
	if n := len(g.ScopedEffects); n != 0 {
		t.Errorf("RestoreFrom kept %d scoped effects, want 0", n)
	}
	if n := len(g.BuiltinReplacements); n != len(snap.BuiltinReplacements) {
		t.Errorf("RestoreFrom lost the built-in registry: %d entries, want %d", n, len(snap.BuiltinReplacements))
	}
}

// TestCloneCopiesScopedReplacements covers the other direction: a
// snapshot taken WHILE a Fog is live must carry it, and sweeping the
// original afterwards must not reach the snapshot's copy.
func TestCloneCopiesScopedReplacements(t *testing.T) {
	g := newActiveGame(t)
	g.WithWriteLock(func() {
		g.PreventCombatDamageThisTurnForEffect(uuid.Nil, uuid.Nil, "Fog")
	})

	snap := g.Clone()
	if len(snap.ScopedEffects) != 1 {
		t.Fatalf("clone dropped the live Fog")
	}

	g.WithWriteLock(func() { g.ClearEndOfTurnScopedStaticsLocked() })
	if len(g.ScopedEffects) != 0 {
		t.Fatalf("setup: the cleanup sweep kept the Fog: %d records", len(g.ScopedEffects))
	}
	if len(snap.ScopedEffects) != 1 {
		t.Error("sweeping the original's registry reached the clone (aliased slice)")
	}
}

// TestRestoreFromForcesLayerRecompute proves a restore leaves the
// layer engine stale (layerVersion > lastResolvedVersion) so the
// next snapshot recomputes `effective` from the restored board
// instead of serving values computed for the pre-undo state.
func TestRestoreFromForcesLayerRecompute(t *testing.T) {
	g := newActiveGame(t)
	// Settle the engine so the pre-restore counters are equal.
	g.ReadSnapshot(func() {})
	snap := g.Clone()

	g.WithWriteLock(func() { g.RestoreFrom(snap) })
	if lv, rv := g.layerVersion.Load(), g.lastResolvedVersion.Load(); lv <= rv {
		t.Errorf("after RestoreFrom: layerVersion=%d lastResolvedVersion=%d; want stale (layerVersion greater)", lv, rv)
	}
	// And the staleness must actually resolve on the next snapshot.
	before := g.LayerRecomputeCountForTest()
	g.ReadSnapshot(func() {})
	if g.LayerRecomputeCountForTest() == before {
		t.Error("snapshot after RestoreFrom did not run a layer recompute")
	}
}
