package game

import (
	"errors"
	"testing"

	"github.com/google/uuid"
)

// #5838: "<player> skips their next combat phase" (Blinding Angel;
// skip_next_combat_phase.go; CR 500.11, 614.10, 614.10a).

func skipNextCombatOf(t *testing.T, g *Game, player uuid.UUID) {
	t.Helper()
	var ok bool
	g.WithWriteLock(func() {
		ok = g.SkipNextCombatPhaseForEffect(player, uuid.Nil, "probe")
	})
	if !ok {
		t.Fatal("SkipNextCombatPhaseForEffect recorded nothing")
	}
}

func nextCombatSkipRecords(g *Game) int {
	n := 0
	for _, e := range g.ScopedEffects {
		if _, ok := nextCombatSkipRecord(e); ok {
			n++
		}
	}
	return n
}

func countStep(steps []Step, want Step) int {
	n := 0
	for _, s := range steps {
		if s == want {
			n++
		}
	}
	return n
}

// finishTurnFrom advances from the cursor to the end of the active
// seat's turn and returns the steps that began, cut at its cleanup.
func finishTurnFrom(t *testing.T, g *Game, since uint64) []Step {
	t.Helper()
	seat := g.Turn.ActiveSeat
	for i := 0; i < 64 && g.Turn.ActiveSeat == seat; i++ {
		if _, err := g.AdvanceStep(); err != nil {
			t.Fatalf("AdvanceStep: %v", err)
		}
	}
	if g.Turn.ActiveSeat == seat {
		t.Fatalf("seat %d's turn never ended", seat)
	}
	steps := stepsBegunSince(g, since)
	for i, s := range steps {
		if s == StepCleanup {
			return steps[:i+1]
		}
	}
	return steps
}

// addCombatAndMain is Relentless Assault in the precombat main phase:
// the turn then has two combat phases, the added one first.
func addCombatAndMain(t *testing.T, g *Game) {
	t.Helper()
	g.WithWriteLock(func() {
		if ids := g.AddPhasesForEffect(uuid.Nil, PhaseAnchor{Kind: AnchorThisMainPhase}, PhaseKindCombat, PhaseKindMain); len(ids) != 2 {
			t.Fatalf("added %v, want two phases", ids)
		}
	})
}

// The skip waits for the player's next combat phase, which is in their
// next turn: the caster's combat is untouched, the player's turn has no
// combat and keeps its main phases and end step, the record is spent, and
// their turn after that has combat again.
func TestNextCombatSkipTakesThatPlayersNextCombatPhase(t *testing.T) {
	g := newFourPlayerActiveGame(t)
	advanceToStepOfSeat(t, g, 0, StepPrecombatMain)
	skipNextCombatOf(t, g, g.Seats[1].ID)
	if n := g.NextCombatPhaseSkipsForLocked(g.Seats[1].ID); n != 1 {
		t.Fatalf("%d skips owed, want 1", n)
	}
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
	if n := nextCombatSkipRecords(g); n != 0 {
		t.Errorf("%d records left after the skipped phase, want 0", n)
	}
	if steps := walkTurnOf(t, g, 1); !anyCombatStep(steps) {
		t.Errorf("seat 1's turn after that had no combat: %v", steps)
	}
}

// One skip is one combat phase: a turn with two loses only the first,
// and the second begins as usual.
func TestNextCombatSkipTakesOnlyTheFirstOfTwoCombatPhases(t *testing.T) {
	g := newFourPlayerActiveGame(t)
	skipNextCombatOf(t, g, g.Seats[1].ID)
	advanceToStepOfSeat(t, g, 1, StepPrecombatMain)
	addCombatAndMain(t, g)
	steps := finishTurnFrom(t, g, eventHighWater(g))
	if n := countStep(steps, StepBeginCombat); n != 1 {
		t.Fatalf("%d beginning of combat steps began, want 1: %v", n, steps)
	}
	if n := countStep(steps, StepDeclareAttackers); n != 1 {
		t.Fatalf("%d declare attackers steps began, want 1: %v", n, steps)
	}
	if n := nextCombatSkipRecords(g); n != 0 {
		t.Errorf("%d records left, want 0", n)
	}
}

// Two skips are two combat phases (CR 614.10a), the two of one turn
// here; the turn after has combat again.
func TestTwoNextCombatSkipsTakeTwoCombatPhases(t *testing.T) {
	g := newFourPlayerActiveGame(t)
	skipNextCombatOf(t, g, g.Seats[1].ID)
	skipNextCombatOf(t, g, g.Seats[1].ID)
	if n := g.NextCombatPhaseSkipsForLocked(g.Seats[1].ID); n != 2 {
		t.Fatalf("%d skips owed, want 2", n)
	}
	advanceToStepOfSeat(t, g, 1, StepPrecombatMain)
	addCombatAndMain(t, g)
	if steps := finishTurnFrom(t, g, eventHighWater(g)); anyCombatStep(steps) {
		t.Fatalf("a combat phase began in the turn with two skips: %v", steps)
	}
	if n := nextCombatSkipRecords(g); n != 0 {
		t.Errorf("%d records left, want 0", n)
	}
	if steps := walkTurnOf(t, g, 1); !anyCombatStep(steps) {
		t.Errorf("seat 1's next turn had no combat: %v", steps)
	}
}

// Two skips across turns: one combat phase per turn, then combat again.
func TestTwoNextCombatSkipsTakeTheCombatOfTwoTurns(t *testing.T) {
	g := newFourPlayerActiveGame(t)
	skipNextCombatOf(t, g, g.Seats[1].ID)
	skipNextCombatOf(t, g, g.Seats[1].ID)
	if steps := walkTurnOf(t, g, 1); anyCombatStep(steps) {
		t.Fatalf("first turn had combat: %v", steps)
	}
	if n := g.NextCombatPhaseSkipsForLocked(g.Seats[1].ID); n != 1 {
		t.Fatalf("%d skips owed after one turn, want 1", n)
	}
	if steps := walkTurnOf(t, g, 1); anyCombatStep(steps) {
		t.Fatalf("second turn had combat: %v", steps)
	}
	if steps := walkTurnOf(t, g, 1); !anyCombatStep(steps) {
		t.Fatalf("third turn had no combat: %v", steps)
	}
}

// A combat phase that has begun can't be skipped (CR 614.10): a skip made
// during the player's own combat waits for their next one.
func TestANextCombatSkipMadeInThatPlayersCombatWaits(t *testing.T) {
	g := newFourPlayerActiveGame(t)
	advanceToStepOfSeat(t, g, 0, StepBeginCombat)
	skipNextCombatOf(t, g, g.Seats[0].ID)
	since := eventHighWater(g)
	advanceToStepOfSeat(t, g, 0, StepPostcombatMain)
	if !stepsContain(stepsBegunSince(g, since), StepEndCombat) {
		t.Fatal("the combat phase in progress was cut short")
	}
	if n := nextCombatSkipRecords(g); n != 1 {
		t.Fatalf("%d records after the combat in progress, want 1", n)
	}
	if steps := walkTurnOf(t, g, 0); anyCombatStep(steps) {
		t.Fatalf("seat 0's next turn had combat: %v", steps)
	}
}

// A skip made in the player's own turn before combat takes that turn's
// combat phase: it is their next one.
func TestANextCombatSkipMadeBeforeThatPlayersCombatTakesIt(t *testing.T) {
	g := newFourPlayerActiveGame(t)
	advanceToStepOfSeat(t, g, 0, StepPrecombatMain)
	skipNextCombatOf(t, g, g.Seats[0].ID)
	steps := finishTurnFrom(t, g, eventHighWater(g))
	if anyCombatStep(steps) {
		t.Fatalf("the turn in progress kept its combat: %v", steps)
	}
	if !stepsContain(steps, StepPostcombatMain) {
		t.Errorf("the turn lost its postcombat main phase: %v", steps)
	}
}

// A combat phase another effect skips doesn't use the record: it waits
// for the first one that isn't skipped (CR 614.10a).
func TestANextCombatSkipWaitsPastATurnWhoseCombatIsSkipped(t *testing.T) {
	g := newFourPlayerActiveGame(t)
	skipCombatOfNextTurn(t, g, g.Seats[1].ID)
	skipNextCombatOf(t, g, g.Seats[1].ID)
	if steps := walkTurnOf(t, g, 1); anyCombatStep(steps) {
		t.Fatalf("first turn had combat: %v", steps)
	}
	if n := nextCombatSkipRecords(g); n != 1 {
		t.Fatalf("%d next-combat records after a turn skipped whole, want 1", n)
	}
	if steps := walkTurnOf(t, g, 1); anyCombatStep(steps) {
		t.Fatalf("second turn had combat: %v", steps)
	}
	if steps := walkTurnOf(t, g, 1); !anyCombatStep(steps) {
		t.Fatalf("third turn had no combat: %v", steps)
	}
}

// A player who has left writes nothing, owes nothing, and a waiting
// record of a player who leaves is dropped as the next turn begins
// (CR 800.4k).
func TestANextCombatSkipForAPlayerWhoHasLeft(t *testing.T) {
	g := newFourPlayerActiveGame(t)
	g.Seats[3].Eliminated = true
	g.WithWriteLock(func() {
		if g.SkipNextCombatPhaseForEffect(g.Seats[3].ID, uuid.Nil, "probe") {
			t.Error("a skip was recorded for a player who has left")
		}
	})
	skipNextCombatOf(t, g, g.Seats[2].ID)
	g.Seats[2].Eliminated = true
	if n := g.NextCombatPhaseSkipsForLocked(g.Seats[2].ID); n != 0 {
		t.Errorf("a player who has left owes %d skips", n)
	}
	passTurnTo(t, g, 1)
	if n := nextCombatSkipRecords(g); n != 0 {
		t.Errorf("%d records left for a player who has left, want 0", n)
	}
}

// The record is plain data: a restore point keeps it, and an older binary
// refuses the file rather than give the combat back.
func TestNextCombatSkipsSurviveASnapshot(t *testing.T) {
	g := newFourPlayerActiveGame(t)
	skipNextCombatOf(t, g, g.Seats[1].ID)
	_, restored := roundTrip(t, g)
	if n := nextCombatSkipRecords(restored); n != 1 {
		t.Fatalf("restored game has %d records, want 1", n)
	}
	if steps := walkTurnOf(t, restored, 1); anyCombatStep(steps) {
		t.Fatalf("the restored skip did not hold: %v", steps)
	}

	snap := g.CaptureSnapshot()
	spec, ok := modKinds[ModSkipNextCombatPhase]
	if !ok {
		t.Fatal("skipNextCombatPhase is not in the mod-kind table")
	}
	delete(modKinds, ModSkipNextCombatPhase)
	defer func() { modKinds[ModSkipNextCombatPhase] = spec }()
	if _, err := snap.Restore(); !errors.Is(err, ErrUnknownEffectKey) {
		t.Errorf("Restore without skipNextCombatPhase = %v, want ErrUnknownEffectKey", err)
	}
}

func TestASkipNextCombatPhaseModThatNamesNoPlayerIsRefused(t *testing.T) {
	if skipNextCombatPhaseModProblem(Mod{Kind: ModSkipNextCombatPhase}) == "" {
		t.Error("a mod naming no player was accepted")
	}
	if skipNextCombatPhaseModProblem(Mod{Kind: ModSkipNextCombatPhase, Player: uuid.New(), Amount: 3}) == "" {
		t.Error("a mod carrying an amount was accepted")
	}
	if p := skipNextCombatPhaseModProblem(Mod{Kind: ModSkipNextCombatPhase, Player: uuid.New()}); p != "" {
		t.Errorf("a sound mod refused: %s", p)
	}
	g := newActiveGame(t)
	skipNextCombatOf(t, g, g.Seats[1].ID)
	snap := g.CaptureSnapshot()
	snap.ScopedEffects[0].Mods[0].Player = uuid.Nil
	if _, err := snap.RestoreStrict(); !errors.Is(err, ErrUnknownEffectKey) {
		t.Errorf("a restore point naming no player = %v, want ErrUnknownEffectKey", err)
	}
}
