package game

import (
	"testing"

	"github.com/google/uuid"
)

// scoped_replacements_test.go pins the engine half of ADR 0041 P8 for
// replacement effects: the ID scheme, IDs that survive the registry
// shrinking under an open CR 616 prompt, and the Seq counter across a
// restore. The card behaviour is pinned in cards/effects.

func TestScopedReplacementIDRoundTrips(t *testing.T) {
	for _, tc := range []struct {
		seq int64
		mod int
	}{{1, 0}, {1, 7}, {42, 3}, {1 << 40, 0}} {
		id, ok := scopedReplacementID(tc.seq, tc.mod)
		if !ok {
			t.Fatalf("scopedReplacementID(%d, %d) not representable", tc.seq, tc.mod)
		}
		if id < scopedReplacementIDBase || id >= testReplacementIDBase {
			t.Errorf("scopedReplacementID(%d, %d) = %d, outside its range", tc.seq, tc.mod, id)
		}
		seq, mod, ok := decodeScopedReplacementID(id)
		if !ok || seq != tc.seq || mod != tc.mod {
			t.Errorf("decode(%d) = (%d, %d, %v), want (%d, %d, true)", id, seq, mod, ok, tc.seq, tc.mod)
		}
	}
	for _, bad := range []struct {
		seq int64
		mod int
	}{{0, 0}, {-1, 0}, {1, scopedReplacementModStride}, {1, -1}, {1 << 62, 0}} {
		if _, ok := scopedReplacementID(bad.seq, bad.mod); ok {
			t.Errorf("scopedReplacementID(%d, %d) accepted", bad.seq, bad.mod)
		}
	}
	for _, id := range []ReplacementEffectID{scopedReplacementIDBase, testReplacementIDBase, 1} {
		if _, _, ok := decodeScopedReplacementID(id); ok {
			t.Errorf("decodeScopedReplacementID(%d) accepted an ID outside the scoped range", id)
		}
	}
}

// The reason the ID comes from Seq and not a position: a CR 616 prompt
// holds the IDs of the effects it lists while the registry shrinks
// under it (here a record ahead of them is swept), and the answer must
// still name the effects the prompt showed.
func TestScopedReplacementIDsSurviveAShrinkingRegistry(t *testing.T) {
	g := newActiveGame(t)
	victim := g.Seats[1]
	life := victim.Life
	g.WithWriteLock(func() {
		g.PreventCombatDamageThisTurnForEffect(uuid.Nil, uuid.Nil, "Fog")
		g.PreventNextDamageThisTurnForEffect(uuid.Nil, victim.ID, 2, false, "two-point shield")
		g.PreventNextDamageThisTurnForEffect(uuid.Nil, victim.ID, 3, false, "three-point shield")
		if err := g.DealDamageToPlayerForEffect(uuid.New(), victim.ID, 4); err != nil {
			t.Fatalf("deal: %v", err)
		}
	})
	prompt := pendingOfKind(t, g, PendingChoiceReplacementOrder)
	if len(prompt.ReplacementEffectIDs) != 2 {
		t.Fatalf("prompt lists %d effects, want the two shields", len(prompt.ReplacementEffectIDs))
	}
	byLabel := map[string]ReplacementEffectID{}
	g.ReadSnapshot(func() {
		for _, id := range prompt.ReplacementEffectIDs {
			label, _ := g.ReplacementOptionMetaForEffect(id)
			byLabel[label] = id
		}
	})
	if byLabel["two-point shield"] == 0 || byLabel["three-point shield"] == 0 {
		t.Fatalf("prompt labels do not resolve: %v", byLabel)
	}

	// The Fog, ahead of both shields, leaves the registry.
	g.WithWriteLock(func() {
		g.ScopedEffects = append([]ScopedEffect(nil), g.ScopedEffects[1:]...)
	})
	g.ReadSnapshot(func() {
		if label, _ := g.ReplacementOptionMetaForEffect(byLabel["three-point shield"]); label != "three-point shield" {
			t.Errorf("after the registry shrank the three-point ID names %q", label)
		}
	})

	if err := g.ResolveReplacementOrder(prompt.ID, prompt.Chooser,
		[]ReplacementEffectID{byLabel["three-point shield"], byLabel["two-point shield"]}); err != nil {
		t.Fatalf("ResolveReplacementOrder: %v", err)
	}
	if victim.Life != life {
		t.Errorf("life %d → %d: 4 damage into a 3- then a 2-point shield is all prevented", life, victim.Life)
	}
	g.ReadSnapshot(func() {
		if len(g.ScopedEffects) != 1 {
			t.Fatalf("%d records left, want the two-point shield with 1 charge", len(g.ScopedEffects))
		}
		e := g.ScopedEffects[0]
		if e.Label != "two-point shield" || e.Mods[0].Amount != 1 {
			t.Errorf("left %q with charge %d, want the two-point shield with 1", e.Label, e.Mods[0].Amount)
		}
	})
}

// A layer-only record carries no Seq — it is never named, and leaving
// it zero keeps it the bytes an earlier v7 binary wrote — and a
// replacement record takes the next one.
func TestOnlyANamedRecordTakesASeq(t *testing.T) {
	g := newActiveGame(t)
	id := pushScopedTestCreature(g, g.Seats[0].ID, 2, 2)
	registerScopedEffectForTest(t, g, id, []Mod{ModifyPTMod(1, 1)}, IndefiniteDuration())
	g.WithWriteLock(func() { g.PreventCombatDamageThisTurnForEffect(uuid.Nil, uuid.Nil, "Fog") })
	g.ReadSnapshot(func() {
		if g.ScopedEffects[0].Seq != 0 {
			t.Errorf("a layer-only record took Seq %d", g.ScopedEffects[0].Seq)
		}
		if g.ScopedEffects[1].Seq != 1 {
			t.Errorf("the Fog took Seq %d, want 1", g.ScopedEffects[1].Seq)
		}
	})
}

// Restore sets the counter to the largest Seq carried, so a record
// registered after a restore never takes a Seq a live record holds.
func TestRestoreRebuildsTheScopedEffectSeq(t *testing.T) {
	g := newActiveGame(t)
	g.WithWriteLock(func() {
		g.PreventCombatDamageThisTurnForEffect(uuid.Nil, uuid.Nil, "Fog")
		g.PreventCombatDamageThisTurnForEffect(uuid.Nil, g.Seats[1].ID, "Druid's Deliverance")
	})
	snap := g.CaptureSnapshot()
	if !snap.Restorable() {
		t.Fatalf("a Fog is data now; the game should be a restore point: %+v", snap.Continuations)
	}
	restored, err := snap.RestoreStrict()
	if err != nil {
		t.Fatalf("RestoreStrict: %v", err)
	}
	restored.WithWriteLock(func() {
		restored.PreventCombatDamageThisTurnForEffect(uuid.Nil, uuid.Nil, "another Fog")
	})
	seen := map[int64]bool{}
	restored.ReadSnapshot(func() {
		for _, e := range restored.ScopedEffects {
			if seen[e.Seq] {
				t.Errorf("Seq %d is held by two records", e.Seq)
			}
			seen[e.Seq] = true
		}
	})
	if !seen[3] {
		t.Errorf("the record registered after the restore did not take Seq 3: %v", seen)
	}
}

// exileInsteadOfLeaving: the pinned object's battlefield exit goes to
// exile, and the record lasts until that object is gone — not until
// cleanup (#1591).
func TestExileInsteadOfLeavingLastsAsLongAsItsObject(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0]
	id := pushScopedTestCreature(g, me.ID, 2, 2)
	g.WithWriteLock(func() {
		if !g.ExileInsteadOfLeavingBattlefieldForEffect(uuid.Nil, id, me.ID, "redirect") {
			t.Fatal("setup: the redirect registered nothing")
		}
		g.ClearEndOfTurnScopedStaticsLocked()
	})
	if len(g.ScopedEffects) != 1 {
		t.Fatalf("the cleanup sweep ended an indefinite redirect: %d records", len(g.ScopedEffects))
	}
	g.WithWriteLock(func() {
		_ = g.DestroyPermanentForEffect(id)
		g.ClearExpiredScopedStaticsLocked()
	})
	if me.Graveyard.Contains(id) || !g.Exile.Contains(id) {
		t.Error("the destroyed permanent was not exiled instead")
	}
	if len(g.ScopedEffects) != 0 {
		t.Errorf("the redirect outlived its object: %d records", len(g.ScopedEffects))
	}
}

// exileInsteadOfGraveyard: a permanent its controller controls is
// exiled instead of dying, and the named body is scheduled for it.
func TestExileInsteadOfGraveyardSchedulesItsBody(t *testing.T) {
	g := newActiveGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	mine := pushScopedTestCreature(g, me.ID, 2, 2)
	theirs := pushScopedTestCreature(g, opp.ID, 2, 2)
	body := testBody(func(*Game, *StackItem) error { return nil })
	g.WithWriteLock(func() {
		if !g.ExileInsteadOfGraveyardThisTurnForEffect(uuid.Nil, me.ID, body, "exile instead") {
			t.Fatal("setup: registered nothing")
		}
		_ = g.DestroyPermanentForEffect(mine)
		_ = g.DestroyPermanentForEffect(theirs)
	})
	if !g.Exile.Contains(mine) || me.Graveyard.Contains(mine) {
		t.Error("the controller's permanent was not exiled instead")
	}
	if !opp.Graveyard.Contains(theirs) {
		t.Error("an opponent's permanent is not covered")
	}
	if len(g.DelayedTriggers) != 1 || g.DelayedTriggers[0].Body.Key() != body.Key() {
		t.Fatalf("delayed triggers = %d, want one naming the registered body", len(g.DelayedTriggers))
	}
}

// A replacement mod is not a layer operation: the layer adapter builds
// nothing for it.
func TestTheLayerAdapterSkipsReplacementMods(t *testing.T) {
	recs := []ScopedEffect{{
		Scope: ScopeGame, Mods: []Mod{{Kind: ModPreventCombatDamage}}, Seq: 1,
	}}
	if got := adaptScopedEffects(recs); len(got) != 0 {
		t.Errorf("the layer adapter built %d continuous effects from a replacement record", len(got))
	}
}
