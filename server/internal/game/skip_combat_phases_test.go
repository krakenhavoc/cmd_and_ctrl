package game

import (
	"errors"
	"testing"

	"github.com/google/uuid"
)

// S79 (#5075): "<player> skips all combat phases of their next turn"
// (skip_combat_phases.go; CR 500.11, 614.10, 614.10a).

func skipCombatOfNextTurn(t *testing.T, g *Game, player uuid.UUID) {
	t.Helper()
	var ok bool
	g.WithWriteLock(func() {
		ok = g.SkipCombatPhasesOfNextTurnForEffect(player, uuid.Nil, "probe")
	})
	if !ok {
		t.Fatal("SkipCombatPhasesOfNextTurnForEffect recorded nothing")
	}
}

func combatSkipRecords(g *Game) int {
	n := 0
	for _, e := range g.ScopedEffects {
		if _, ok := combatSkipRecord(e); ok {
			n++
		}
	}
	return n
}

// walkTurnOf advances from wherever the cursor is to the start of
// `seat`'s next turn, then through that whole turn, and returns the
// steps that BEGAN during it.
func walkTurnOf(t *testing.T, g *Game, seat int) []Step {
	t.Helper()
	advanceToStepOfSeat(t, g, seat, StepUpkeep)
	since := eventHighWater(g)
	for i := 0; i < 64 && g.Turn.ActiveSeat == seat; i++ {
		if _, err := g.AdvanceStep(); err != nil {
			t.Fatalf("AdvanceStep in seat %d's turn: %v", seat, err)
		}
	}
	if g.Turn.ActiveSeat == seat {
		t.Fatalf("seat %d's turn never ended", seat)
	}
	out := stepsBegunSince(g, since)
	// The steps after the turn changed belong to the next seat; cut at
	// this turn's cleanup step.
	for i, s := range out {
		if s == StepCleanup {
			return out[:i+1]
		}
	}
	return out
}

func anyCombatStep(steps []Step) bool {
	for _, s := range steps {
		if PhaseOf(s) == PhaseCombat {
			return true
		}
	}
	return false
}

// The skip is for the target's NEXT turn: the turn in progress (the
// caster's) keeps its combat, the target's next turn has none and still
// has its main phases and end step, and the turn after that has combat
// again. The record is spent.
func TestCombatSkipTakesEveryCombatPhaseOfThatPlayersNextTurn(t *testing.T) {
	g := newFourPlayerActiveGame(t)
	target := g.Seats[1]
	advanceToStepOfSeat(t, g, 0, StepPrecombatMain)
	skipCombatOfNextTurn(t, g, target.ID)
	since := eventHighWater(g)
	advanceToStepOfSeat(t, g, 0, StepPostcombatMain)
	if !stepsContain(stepsBegunSince(g, since), StepBeginCombat) {
		t.Fatal("the caster's own combat was skipped")
	}
	steps := walkTurnOf(t, g, 1)
	if anyCombatStep(steps) {
		t.Fatalf("seat 1's next turn had a combat step: %v", steps)
	}
	for _, want := range []Step{StepPrecombatMain, StepPostcombatMain, StepEnd} {
		if !stepsContain(steps, want) {
			t.Errorf("seat 1's turn lost its %s: %v", want, steps)
		}
	}
	if steps := walkTurnOf(t, g, 2); !anyCombatStep(steps) {
		t.Errorf("seat 2's turn had no combat: %v", steps)
	}
	if n := combatSkipRecords(g); n != 0 {
		t.Errorf("%d records left after the skipped turn, want 0", n)
	}
	if steps := walkTurnOf(t, g, 1); !anyCombatStep(steps) {
		t.Errorf("seat 1's turn after that had no combat: %v", steps)
	}
}

// "All combat phases": one added during the skipped turn is skipped too.
func TestCombatSkipTakesACombatPhaseAddedThatTurn(t *testing.T) {
	g := newFourPlayerActiveGame(t)
	skipCombatOfNextTurn(t, g, g.Seats[1].ID)
	advanceToStepOfSeat(t, g, 1, StepPrecombatMain)
	if g.CombatPhaseSkipsForLocked(g.Seats[1].ID) != 1 {
		t.Fatal("the armed skip is not reported for the turn in progress")
	}
	since := eventHighWater(g)
	g.WithWriteLock(func() {
		if ids := g.AddPhasesForEffect(uuid.Nil, PhaseAnchor{Kind: AnchorThisMainPhase}, PhaseKindCombat, PhaseKindMain); len(ids) != 2 {
			t.Fatalf("added %v, want two phases", ids)
		}
	})
	for i := 0; i < 64 && g.Turn.ActiveSeat == 1; i++ {
		if _, err := g.AdvanceStep(); err != nil {
			t.Fatalf("AdvanceStep: %v", err)
		}
	}
	steps := stepsBegunSince(g, since)
	cut := len(steps)
	for i, s := range steps {
		if s == StepCleanup {
			cut = i + 1
			break
		}
	}
	if anyCombatStep(steps[:cut]) {
		t.Fatalf("an added combat phase of the skipped turn began: %v", steps[:cut])
	}
}

// Two skips are two turns (CR 614.10a): the oldest is satisfied by the
// first turn, the other waits for the next.
func TestTwoCombatSkipsTakeTwoTurns(t *testing.T) {
	g := newFourPlayerActiveGame(t)
	skipCombatOfNextTurn(t, g, g.Seats[1].ID)
	skipCombatOfNextTurn(t, g, g.Seats[1].ID)
	if n := g.CombatPhaseSkipsForLocked(g.Seats[1].ID); n != 2 {
		t.Fatalf("%d skips owed, want 2", n)
	}
	if steps := walkTurnOf(t, g, 1); anyCombatStep(steps) {
		t.Fatalf("first turn had combat: %v", steps)
	}
	if steps := walkTurnOf(t, g, 1); anyCombatStep(steps) {
		t.Fatalf("second turn had combat: %v", steps)
	}
	if steps := walkTurnOf(t, g, 1); !anyCombatStep(steps) {
		t.Fatalf("third turn had no combat: %v", steps)
	}
}

// A skip made during the target's own turn waits for their next one
// (CR 614.10: a phase that has started can't be skipped, and "their next
// turn" is not this one).
func TestACombatSkipMadeInThatPlayersTurnWaitsForTheNext(t *testing.T) {
	g := newFourPlayerActiveGame(t)
	advanceToStepOfSeat(t, g, 0, StepPrecombatMain)
	skipCombatOfNextTurn(t, g, g.Seats[0].ID)
	since := eventHighWater(g)
	advanceToStepOfSeat(t, g, 0, StepPostcombatMain)
	if !stepsContain(stepsBegunSince(g, since), StepBeginCombat) {
		t.Fatal("the turn in progress lost its combat")
	}
	if steps := walkTurnOf(t, g, 0); anyCombatStep(steps) {
		t.Fatalf("seat 0's next turn had combat: %v", steps)
	}
}

// A player who has left writes nothing, and a waiting record of a player
// who leaves is dropped as the next turn begins (CR 800.4k).
func TestACombatSkipForAPlayerWhoHasLeft(t *testing.T) {
	g := newFourPlayerActiveGame(t)
	g.Seats[3].Eliminated = true
	g.WithWriteLock(func() {
		if g.SkipCombatPhasesOfNextTurnForEffect(g.Seats[3].ID, uuid.Nil, "probe") {
			t.Error("a skip was recorded for a player who has left")
		}
	})
	skipCombatOfNextTurn(t, g, g.Seats[2].ID)
	g.Seats[2].Eliminated = true
	if n := g.CombatPhaseSkipsForLocked(g.Seats[2].ID); n != 0 {
		t.Errorf("a player who has left owes %d skips", n)
	}
	passTurnTo(t, g, 1)
	if n := combatSkipRecords(g); n != 0 {
		t.Errorf("%d records left for a player who has left, want 0", n)
	}
}

// The record is plain data: a restore point keeps it, waiting or armed,
// and an older binary refuses the file rather than give the combat back.
func TestCombatSkipsSurviveASnapshot(t *testing.T) {
	g := newFourPlayerActiveGame(t)
	skipCombatOfNextTurn(t, g, g.Seats[1].ID)
	_, restored := roundTrip(t, g)
	if n := combatSkipRecords(restored); n != 1 {
		t.Fatalf("restored game has %d records, want 1", n)
	}
	if steps := walkTurnOf(t, restored, 1); anyCombatStep(steps) {
		t.Fatalf("the restored skip did not hold: %v", steps)
	}

	// Armed, mid-turn: a restore in the skipped turn's main phase, whose
	// plan is rebuilt from the template, still skips the combat.
	g2 := newFourPlayerActiveGame(t)
	skipCombatOfNextTurn(t, g2, g2.Seats[1].ID)
	advanceToStepOfSeat(t, g2, 1, StepPrecombatMain)
	_, mid := roundTrip(t, g2)
	since := eventHighWater(mid)
	advanceToStepOfSeat(t, mid, 1, StepPostcombatMain)
	if anyCombatStep(stepsBegunSince(mid, since)) {
		t.Fatal("a restore in the skipped turn brought its combat back")
	}

	snap := g.CaptureSnapshot()
	spec, ok := modKinds[ModSkipCombatPhases]
	if !ok {
		t.Fatal("skipCombatPhases is not in the mod-kind table")
	}
	delete(modKinds, ModSkipCombatPhases)
	defer func() { modKinds[ModSkipCombatPhases] = spec }()
	if _, err := snap.Restore(); !errors.Is(err, ErrUnknownEffectKey) {
		t.Errorf("Restore without skipCombatPhases = %v, want ErrUnknownEffectKey", err)
	}
}

func TestASkipCombatPhasesModThatNamesNoPlayerIsRefused(t *testing.T) {
	if skipCombatPhasesModProblem(Mod{Kind: ModSkipCombatPhases}) == "" {
		t.Error("a mod naming no player was accepted")
	}
	if skipCombatPhasesModProblem(Mod{Kind: ModSkipCombatPhases, Player: uuid.New(), Amount: -1}) == "" {
		t.Error("a negative turn was accepted")
	}
	if p := skipCombatPhasesModProblem(Mod{Kind: ModSkipCombatPhases, Player: uuid.New()}); p != "" {
		t.Errorf("a sound mod refused: %s", p)
	}
	g := newActiveGame(t)
	skipCombatOfNextTurn(t, g, g.Seats[1].ID)
	snap := g.CaptureSnapshot()
	snap.ScopedEffects[0].Mods[0].Player = uuid.Nil
	if _, err := snap.RestoreStrict(); !errors.Is(err, ErrUnknownEffectKey) {
		t.Errorf("a restore point naming no player = %v, want ErrUnknownEffectKey", err)
	}
}
