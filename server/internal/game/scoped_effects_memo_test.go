package game

import (
	"testing"

	"github.com/google/uuid"
)

// scoped_effects_memo_test.go pins the layer-pass adapter memo (#1558):
// reused while the records are the same records, rebuilt the moment
// they are not — by registration, by the sweep, by an undo that takes
// a record away, by a restore — and never shared between a game and
// its clone.

func adaptedScopedEffects(g *Game) []ContinuousEffect {
	var out []ContinuousEffect
	g.WithWriteLock(func() { out = g.scopedEffectContinuousEffectsLocked() })
	return out
}

func sameBacking(a, b []ContinuousEffect) bool {
	return len(a) > 0 && len(b) > 0 && &a[0] == &b[0]
}

func TestScopedEffectAdapterIsReusedWhileNothingChanged(t *testing.T) {
	g := newActiveGame(t)
	bear := pushScopedTestCreature(g, g.Seats[0].ID, 2, 2)
	registerScopedEffectForTest(t, g, bear, []Mod{ModifyPTMod(1, 1), AddKeywordsMod("flying")}, IndefiniteDuration())

	first := adaptedScopedEffects(g)
	if len(first) != 2 {
		t.Fatalf("%d adapted effects, want one per mod (2)", len(first))
	}
	// An unrelated layer-version bump is not a change to the records.
	g.BumpLayerVersionForTest()
	if again := adaptedScopedEffects(g); !sameBacking(first, again) {
		t.Error("the adapter rebuilt with no change to the records")
	}
	if cap(first) != len(first) {
		t.Errorf("returned slice has spare capacity %d > %d: a caller's append would write into the memo", cap(first), len(first))
	}
}

func TestScopedEffectAdapterRebuildsOnRegistrationAndSweep(t *testing.T) {
	g := newActiveGame(t)
	bear := pushScopedTestCreature(g, g.Seats[0].ID, 2, 2)
	registerScopedEffectForTest(t, g, bear, []Mod{ModifyPTMod(1, 1)}, IndefiniteDuration())
	before := adaptedScopedEffects(g)

	var eot Duration
	g.WithWriteLock(func() { eot = g.UntilEndOfTurnDuration() })
	registerScopedEffectForTest(t, g, bear, []Mod{ModifyPTMod(2, 2)}, eot)
	if after := adaptedScopedEffects(g); len(after) != 2 || sameBacking(before, after) {
		t.Fatalf("after a registration: %d effects (want 2), reused=%v", len(after), sameBacking(before, after))
	}
	if c := scopedEffectChar(t, g, bear); c.Power != 5 || c.Toughness != 5 {
		t.Fatalf("P/T %d/%d, want 5/5 with both records", c.Power, c.Toughness)
	}

	// The cleanup sweep ends the until-end-of-turn record (CR 514.2).
	g.WithWriteLock(func() {
		if g.sweepScopedEffectsLocked(true) {
			g.layerVersion.Add(1)
		}
	})
	if n := len(adaptedScopedEffects(g)); n != 1 {
		t.Errorf("after the sweep: %d adapted effects, want 1", n)
	}
	if c := scopedEffectChar(t, g, bear); c.Power != 3 || c.Toughness != 3 {
		t.Errorf("after the sweep: P/T %d/%d, want 3/3", c.Power, c.Toughness)
	}
}

// TestScopedEffectAdapterSurvivesAnUndoThatRewindsTheVersion is the
// case a layerVersion key gets wrong: an undo stores the snapshot's
// version plus one, and the registration after it can land the game
// back on the very version the memo was built at — with a different
// record in the slot, possibly in the same backing array.
func TestScopedEffectAdapterSurvivesAnUndoThatRewindsTheVersion(t *testing.T) {
	g := newActiveGame(t)
	bear := pushScopedTestCreature(g, g.Seats[0].ID, 2, 2)
	registerScopedEffectForTest(t, g, bear, []Mod{ModifyPTMod(1, 1)}, IndefiniteDuration())
	// Room for the next append in place, so the undo's slice and the
	// live one share a backing array — the worst case.
	g.WithWriteLock(func() {
		grown := make([]ScopedEffect, len(g.ScopedEffects), 4)
		copy(grown, g.ScopedEffects)
		g.ScopedEffects = grown
	})
	undo := g.Clone()
	undo.ScopedEffects = g.ScopedEffects // share the array, as a shallow undo could

	registerScopedEffectForTest(t, g, bear, []Mod{AddSubtypesMod("Orc")}, IndefiniteDuration())
	if c := scopedEffectChar(t, g, bear); !typeListHas(c.Subtypes, "Orc") {
		t.Fatalf("subtypes %v, want Orc before the undo", c.Subtypes)
	}
	built := g.layerVersion.Load()

	g.WithWriteLock(func() { g.RestoreFrom(undo) })
	for g.layerVersion.Load()+1 < built {
		g.layerVersion.Add(1)
	}
	registerScopedEffectForTest(t, g, bear, []Mod{AddSubtypesMod("Elf")}, IndefiniteDuration())
	c := scopedEffectChar(t, g, bear)
	if typeListHas(c.Subtypes, "Orc") {
		t.Errorf("subtypes %v: the undone Orc record came back through the memo", c.Subtypes)
	}
	if !typeListHas(c.Subtypes, "Elf") || c.Power != 3 {
		t.Errorf("subtypes %v, power %d: want the surviving +1/+1 and the new Elf", c.Subtypes, c.Power)
	}
}

func TestScopedEffectAdapterIsRebuiltAfterARestore(t *testing.T) {
	g := newActiveGame(t)
	bear := pushScopedTestCreature(g, g.Seats[0].ID, 2, 2)
	registerScopedEffectForTest(t, g, bear, []Mod{ModifyPTMod(1, 1)}, IndefiniteDuration())
	_ = adaptedScopedEffects(g)

	restored, err := g.CaptureSnapshot().RestoreStrict()
	if err != nil {
		t.Fatalf("RestoreStrict: %v", err)
	}
	if restored.scopedEffectMemo.keys != nil {
		t.Fatal("a restored game arrived with a memo")
	}
	if c := scopedEffectChar(t, restored, bear); c.Power != 3 {
		t.Errorf("restored power %d, want 3", c.Power)
	}
}

func TestACloneDoesNotShareTheAdapterMemo(t *testing.T) {
	g := newActiveGame(t)
	bear := pushScopedTestCreature(g, g.Seats[0].ID, 2, 2)
	registerScopedEffectForTest(t, g, bear, []Mod{ModifyPTMod(1, 1)}, IndefiniteDuration())
	original := adaptedScopedEffects(g)

	clone := g.Clone()
	if clone.scopedEffectMemo.keys != nil || clone.scopedEffectMemo.effects != nil {
		t.Fatal("Clone copied the adapter memo")
	}
	// The clone moves on without touching the original's memo.
	clone.WithWriteLock(func() {
		clone.RegisterScopedEffectForEffect(uuid.Nil, clone.PinnedObjectsLocked(bear),
			[]Mod{ModifyPTMod(5, 5)}, IndefiniteDuration(), "clone only")
	})
	if c := scopedEffectChar(t, clone, bear); c.Power != 8 {
		t.Errorf("clone power %d, want 8", c.Power)
	}
	if again := adaptedScopedEffects(g); !sameBacking(original, again) || len(again) != 1 {
		t.Errorf("the original's memo changed when its clone registered a record (%d effects)", len(again))
	}
	if c := scopedEffectChar(t, g, bear); c.Power != 3 {
		t.Errorf("original power %d, want 3", c.Power)
	}
}
