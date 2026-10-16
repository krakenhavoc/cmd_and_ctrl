package game

import (
	"errors"
	"testing"

	"github.com/google/uuid"
)

// switch_pt_test.go — CR 613.4d's switch (#3996): ModSwitchPT in
// sublayer 7d (SubLayer7E_Switch), after every set, modify and counter.

func switchedPT(t *testing.T, g *Game, id uuid.UUID) (power, toughness, base int) {
	t.Helper()
	g.WithWriteLock(func() {
		g.RecomputeLayersIfStaleLocked()
		c, ok := g.battlefieldCardLocked(id)
		if !ok {
			t.Fatalf("card %s not on battlefield", id)
		}
		_, power, toughness = c.PowerToughness()
		base = c.BasePower()
	})
	return power, toughness, base
}

func setCountersForSwitchTest(g *Game, id uuid.UUID, counters map[string]int) {
	g.WithWriteLock(func() {
		for i := range g.Battlefield.Cards {
			if g.Battlefield.Cards[i].InstanceID == id {
				g.Battlefield.Cards[i].Counters = counters
			}
		}
	})
}

// TestSwitchComesAfterCountersEvenAnUnevenOne: CR 613.4c applies the
// counters before the switch. A 1/4 with a +1/+0 counter is 2/4, which
// switches to 4/2 — not 1/4 switched to 4/1 and the counter added to the
// new power (5/1).
func TestSwitchComesAfterCountersEvenAnUnevenOne(t *testing.T) {
	g := newActiveGame(t)
	id := pushScopedTestCreature(g, g.Seats[0].ID, 1, 4)
	setCountersForSwitchTest(g, id, map[string]int{"+1/+0": 1})
	registerScopedEffectForTest(t, g, id, []Mod{SwitchPTMod()}, IndefiniteDuration())
	if p, tough, _ := switchedPT(t, g, id); p != 4 || tough != 2 {
		t.Fatalf("P/T = %d/%d, want 4/2", p, tough)
	}
	setCountersForSwitchTest(g, id, map[string]int{"+1/+1": 2, "-0/-1": 1})
	// 1/4, +2/+2 and -0/-1: 3/5, switched 5/3.
	if p, tough, _ := switchedPT(t, g, id); p != 5 || tough != 3 {
		t.Fatalf("P/T = %d/%d, want 5/3", p, tough)
	}
}

// TestALaterPumpStillAppliesBeforeTheSwitch: the switch's sublayer comes
// after 7c whatever the timestamps (CR 613.4), so a +2/+0 registered
// after the switch still applies first.
func TestALaterPumpStillAppliesBeforeTheSwitch(t *testing.T) {
	g := newActiveGame(t)
	id := pushScopedTestCreature(g, g.Seats[0].ID, 1, 5)
	registerScopedEffectForTest(t, g, id, []Mod{SwitchPTMod()}, IndefiniteDuration())
	registerScopedEffectForTest(t, g, id, []Mod{ModifyPTMod(2, 0)}, IndefiniteDuration())
	// 1/5 +2/+0 = 3/5, switched 5/3.
	if p, tough, base := switchedPT(t, g, id); p != 5 || tough != 3 || base != 1 {
		t.Fatalf("P/T = %d/%d base %d, want 5/3 base 1", p, tough, base)
	}
}

// TestTwoSwitchesUndoEachOther, counters included.
func TestTwoSwitchesUndoEachOther(t *testing.T) {
	g := newActiveGame(t)
	id := pushScopedTestCreature(g, g.Seats[0].ID, 1, 4)
	setCountersForSwitchTest(g, id, map[string]int{"+1/+0": 1})
	registerScopedEffectForTest(t, g, id, []Mod{SwitchPTMod()}, IndefiniteDuration())
	registerScopedEffectForTest(t, g, id, []Mod{SwitchPTMod()}, IndefiniteDuration())
	if p, tough, _ := switchedPT(t, g, id); p != 2 || tough != 4 {
		t.Fatalf("P/T = %d/%d, want 2/4", p, tough)
	}
}

// TestASwitchEndsAtCleanup: an until-end-of-turn switch is gone the next
// turn.
func TestASwitchEndsAtCleanup(t *testing.T) {
	g := newActiveGame(t)
	id := pushScopedTestCreature(g, g.Seats[0].ID, 1, 4)
	var d Duration
	g.WithWriteLock(func() { d = g.UntilEndOfTurnDuration() })
	registerScopedEffectForTest(t, g, id, []Mod{SwitchPTMod()}, d)
	if p, tough, _ := switchedPT(t, g, id); p != 4 || tough != 1 {
		t.Fatalf("P/T = %d/%d this turn, want 4/1", p, tough)
	}
	passTurn(t, g)
	if p, tough, _ := switchedPT(t, g, id); p != 1 || tough != 4 {
		t.Fatalf("P/T = %d/%d next turn, want 1/4", p, tough)
	}
}

// TestASwitchSurvivesASnapshotAndAnOlderBinaryRefusesIt.
func TestASwitchSurvivesASnapshotAndAnOlderBinaryRefusesIt(t *testing.T) {
	g := newActiveGame(t)
	id := pushScopedTestCreature(g, g.Seats[0].ID, 1, 4)
	registerScopedEffectForTest(t, g, id, []Mod{SwitchPTMod()}, IndefiniteDuration())
	_, restored := roundTrip(t, g)
	if p, tough, _ := switchedPT(t, restored, id); p != 4 || tough != 1 {
		t.Fatalf("restored P/T = %d/%d, want 4/1", p, tough)
	}
	snap := g.CaptureSnapshot()
	spec := modKinds[ModSwitchPT]
	delete(modKinds, ModSwitchPT)
	defer func() { modKinds[ModSwitchPT] = spec }()
	if _, err := snap.Restore(); !errors.Is(err, ErrUnknownEffectKey) {
		t.Errorf("Restore without switchPT = %v, want ErrUnknownEffectKey", err)
	}
}
