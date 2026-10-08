package game

import (
	"testing"

	"github.com/google/uuid"
)

// duration_end_of_combat_test.go — ADR 0108's amendment of 2026-10-07
// (#2027): the UntilEndOfCombat duration. "This combat" and "until end
// of combat" last until the combat PHASE they were made in ends (CR 511.3,
// 724.2d), not until the turn does, and an additional combat phase is a
// different combat.

// combatShield registers an "all combat damage dealt to you" shield for
// seat's player with `d`, and reports whether it was registered.
func combatShield(t *testing.T, g *Game, seat int, d Duration) bool {
	t.Helper()
	var ok bool
	g.WithWriteLock(func() {
		ok = g.PreventDamageFromSourceThisTurnForEffect(DamageShield{
			Controller:    g.Seats[seat].ID,
			ProtectPlayer: g.Seats[seat].ID,
			CombatOnly:    true,
			Duration:      d,
			Label:         "test — prevent combat damage this combat",
		})
	})
	return ok
}

// thisCombat is the duration of the combat phase in progress.
func thisCombat(t *testing.T, g *Game) Duration {
	t.Helper()
	var d Duration
	var ok bool
	g.WithWriteLock(func() { d, ok = g.UntilEndOfCombatDuration() })
	if !ok {
		t.Fatalf("UntilEndOfCombatDuration refused during %s", g.Turn.Step)
	}
	return d
}

func scopedEffectCount(g *Game) int {
	n := 0
	g.ReadSnapshot(func() { n = len(g.ScopedEffects) })
	return n
}

func TestUntilEndOfCombatNeedsACombatPhase(t *testing.T) {
	g := newFourPlayerActiveGame(t)
	advanceToStepOfSeat(t, g, 0, StepPrecombatMain)
	var ok bool
	g.WithWriteLock(func() { _, ok = g.UntilEndOfCombatDuration() })
	if ok {
		t.Fatal("a duration was made in the precombat main phase: there is no combat for it to last through")
	}
	if combatShield(t, g, 0, Duration{}) != true {
		t.Fatal("control: a default-duration shield was not registered")
	}
	advanceToStepOfSeat(t, g, 0, StepDeclareAttackers)
	d := thisCombat(t, g)
	if d.Kind != UntilEndOfCombat || d.CombatTurn != g.Turn.Seq || d.CombatPhase != g.Turn.PhaseID {
		t.Fatalf("duration = %+v, want the stamp of this turn %d, phase %d", d, g.Turn.Seq, g.Turn.PhaseID)
	}
}

// The effect lasts through every step of the combat phase, including the
// end of combat step, and is gone the moment that step ends (CR 511.3) —
// in the same turn, in the postcombat main phase.
func TestUntilEndOfCombatEndsWhenTheEndOfCombatStepEnds(t *testing.T) {
	g := newFourPlayerActiveGame(t)
	advanceToStepOfSeat(t, g, 0, StepBeginCombat)
	d := thisCombat(t, g)
	if !combatShield(t, g, 1, d) {
		t.Fatal("the shield was not registered")
	}
	for _, step := range []Step{StepDeclareAttackers, StepDeclareBlockers, StepCombatDamage, StepEndCombat} {
		advanceToStepOfSeat(t, g, 0, step)
		if n := scopedEffectCount(g); n != 1 {
			t.Fatalf("%d records in %s, want the shield to last the whole combat", n, step)
		}
	}
	advanceToStepOfSeat(t, g, 0, StepPostcombatMain)
	if n := scopedEffectCount(g); n != 0 {
		t.Fatalf("%d records in the postcombat main phase, want the shield gone", n)
	}
}

// An effect that ends the combat phase (CR 724.2d) ends it too.
func TestUntilEndOfCombatEndsWhenAnEffectEndsTheCombatPhase(t *testing.T) {
	g := newFourPlayerActiveGame(t)
	advanceToStepOfSeat(t, g, 0, StepDeclareAttackers)
	if !combatShield(t, g, 1, thisCombat(t, g)) {
		t.Fatal("the shield was not registered")
	}
	g.WithWriteLock(func() { g.EndCombatPhaseForEffect() })
	if g.Turn.Step != StepPostcombatMain {
		t.Fatalf("step = %s, want the postcombat main phase", g.Turn.Step)
	}
	if n := scopedEffectCount(g); n != 0 {
		t.Fatalf("%d records after the combat phase was ended, want none", n)
	}
}

// "This combat" is the phase instance. With an additional combat phase
// planned straight after the first, the shield made in the first one is
// over by the time the second begins, though both are combat phases of
// the same turn.
func TestUntilEndOfCombatIsOneCombatPhaseNotTheTurn(t *testing.T) {
	g := newFourPlayerActiveGame(t)
	advanceToStepOfSeat(t, g, 0, StepBeginCombat)
	first := g.Turn.PhaseID
	g.WithWriteLock(func() {
		if ids := g.AddPhasesForEffect(uuid.Nil, PhaseAnchor{Kind: AnchorThisPhase}, PhaseKindCombat); len(ids) != 1 {
			t.Fatalf("AddPhasesForEffect added %v, want one combat phase", ids)
		}
	})
	if !combatShield(t, g, 1, thisCombat(t, g)) {
		t.Fatal("the shield was not registered")
	}
	advanceToStepOfSeat(t, g, 0, StepEndCombat)
	if n := scopedEffectCount(g); n != 1 {
		t.Fatalf("%d records at the end of the first combat, want 1", n)
	}
	for i := 0; i < 20 && !(g.Turn.PhaseID != first && PhaseOf(g.Turn.Step) == PhaseCombat); i++ {
		if _, err := g.AdvanceStep(); err != nil {
			t.Fatal(err)
		}
	}
	if g.Turn.PhaseID == first || PhaseOf(g.Turn.Step) != PhaseCombat {
		t.Fatalf("never reached the additional combat phase (at %s, phase %d)", g.Turn.Step, g.Turn.PhaseID)
	}
	if n := scopedEffectCount(g); n != 0 {
		t.Fatalf("%d records in the additional combat phase, want the first combat's shield gone", n)
	}
}

// A turn that ends without a real end of combat step (cleanup is reached
// some other way) cannot keep the effect past its turn.
func TestUntilEndOfCombatDoesNotSurviveTheTurn(t *testing.T) {
	g := newFourPlayerActiveGame(t)
	advanceToStepOfSeat(t, g, 0, StepDeclareAttackers)
	d := thisCombat(t, g)
	advanceToStepOfSeat(t, g, 1, StepDeclareAttackers)
	var expired bool
	g.WithWriteLock(func() { expired = g.durationExpiredLocked(d, false) })
	if !expired {
		t.Fatal("a combat of the previous turn is still running in the next turn's combat")
	}
}

// Problem refuses the combat stamp on any other kind, and the kind is
// known to this binary.
func TestUntilEndOfCombatDurationIsValidated(t *testing.T) {
	if !UntilEndOfCombat.Known() {
		t.Fatal("UntilEndOfCombat is not Known — is it after durationKindEnd?")
	}
	if p := (Duration{Kind: UntilEndOfCombat, CombatTurn: 3, CombatPhase: 3}).Problem(); p != "" {
		t.Errorf("a stamped combat duration is refused: %s", p)
	}
	for _, d := range []Duration{
		{Kind: UntilEndOfTurn, CombatTurn: 1},
		{Kind: UntilYourNextTurn, CombatPhase: 3},
		{Kind: ForAsLongAs, CombatTurn: 1, CombatPhase: 3},
		{Kind: Indefinite, CombatPhase: 3},
	} {
		if d.Problem() == "" {
			t.Errorf("%v with a combat stamp was accepted", d.Kind)
		}
	}
	if got := UntilEndOfCombat.String(); got != "until end of combat" {
		t.Errorf("String = %q", got)
	}
}

// A shield takes a stated duration, and refuses two of them (or one that
// restore would refuse) rather than choosing silently.
func TestDamageShieldDurationIsRefusedBesideAnotherOrWhenBroken(t *testing.T) {
	g := newFourPlayerActiveGame(t)
	advanceToStepOfSeat(t, g, 0, StepDeclareAttackers)
	d := thisCombat(t, g)
	var both, broken bool
	g.WithWriteLock(func() {
		both = g.PreventDamageFromSourceThisTurnForEffect(DamageShield{
			Controller: g.Seats[0].ID, ProtectPlayer: g.Seats[0].ID,
			UntilYourNextTurn: true, Duration: d,
		})
		broken = g.PreventDamageFromSourceThisTurnForEffect(DamageShield{
			Controller: g.Seats[0].ID, ProtectPlayer: g.Seats[0].ID,
			Duration: Duration{Kind: UntilEndOfTurn, CombatTurn: 1},
		})
	})
	if both || broken || scopedEffectCount(g) != 0 {
		t.Fatalf("registered with two durations (%v) or a broken one (%v), %d records", both, broken, scopedEffectCount(g))
	}
}

// A restore point written mid-combat carries the stamp, restores, and
// the restored effect ends with the combat like the original.
func TestUntilEndOfCombatSurvivesASnapshotAndClone(t *testing.T) {
	g := newFourPlayerActiveGame(t)
	advanceToStepOfSeat(t, g, 0, StepDeclareAttackers)
	if !combatShield(t, g, 1, thisCombat(t, g)) {
		t.Fatal("the shield was not registered")
	}
	clone := g.Clone()
	restored, err := throughJSON(t, g.CaptureSnapshot()).RestoreStrict()
	if err != nil {
		t.Fatalf("RestoreStrict: %v", err)
	}
	for name, game := range map[string]*Game{"clone": clone, "restored": restored} {
		if n := scopedEffectCount(game); n != 1 {
			t.Fatalf("%s holds %d records, want 1", name, n)
		}
		var d Duration
		game.ReadSnapshot(func() { d = game.ScopedEffects[0].Duration })
		if d.Kind != UntilEndOfCombat || d.CombatPhase == 0 {
			t.Fatalf("%s lost the stamp: %+v", name, d)
		}
		advanceToStepOfSeat(t, game, 0, StepPostcombatMain)
		if n := scopedEffectCount(game); n != 0 {
			t.Errorf("%s still holds %d records after the combat", name, n)
		}
	}
	if n := scopedEffectCount(g); n != 1 {
		t.Errorf("advancing the copies changed the original: %d records", n)
	}
}
