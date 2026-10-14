package game

import (
	"errors"
	"testing"

	"github.com/google/uuid"
)

// turn_skip_test.go pins CR 614.10 "skip your next turn" (ADR 0188
// Decision 1, #3559): a skip debt is a ScopedEffect record the rotation
// seam spends as the player's next turn would begin. Each test drives
// the real seam through PassTurn and reads who is active, never the
// record alone.

// skipTurns writes a debt of n turns for `seat` under the write lock.
func skipTurns(t *testing.T, g *Game, seat int, source uuid.UUID, n int) {
	t.Helper()
	ok := false
	g.WithWriteLock(func() {
		ok = g.SkipNextTurnsForEffect(g.Seats[seat].ID, source, n, "probe — skip")
	})
	if !ok {
		t.Fatalf("SkipNextTurnsForEffect(seat %d, %d) wrote nothing", seat, n)
	}
}

// walkTurns passes the turn len(want) times and compares the seat and
// extra flag of each turn that begins.
func walkTurns(t *testing.T, g *Game, want []turnShape) {
	t.Helper()
	for i, w := range want {
		passTurn(t, g)
		if got := shapeOf(g); got != w {
			t.Fatalf("turn %d: %+v, want %+v", i+1, got, w)
		}
	}
}

// TestASkippedNormalTurnIsPassedOver is the slice's first acceptance
// test: four seats, seat 1 owes one skip, and the order is 0, 2, 3, 0, 1.
// Seat 1's TurnsBegun does not move on the skip, Seq goes up only on
// turns that begin, and Round goes up on the wrap.
func TestASkippedNormalTurnIsPassedOver(t *testing.T) {
	g := newFourPlayerActiveGame(t)
	if g.Turn.ActiveSeat != 0 || g.Turn.Round != 1 {
		t.Fatalf("setup: %+v", shapeOf(g))
	}
	skipTurns(t, g, 1, uuid.Nil, 1)
	begun := g.Seats[1].TurnsBegun
	seq := g.Turn.Seq

	want := []turnShape{
		{seat: 2, round: 1},
		{seat: 3, round: 1},
		{seat: 0, round: 2},
		{seat: 1, round: 2},
	}
	for i, w := range want {
		passTurn(t, g)
		if got := shapeOf(g); got != w {
			t.Fatalf("turn %d: %+v, want %+v", i+1, got, w)
		}
		if g.Turn.Seq != seq+i+1 {
			t.Errorf("turn %d: Seq %d, want %d (one per turn that begins, not per seat passed)", i+1, g.Turn.Seq, seq+i+1)
		}
		if i == 0 && g.Seats[1].TurnsBegun != begun {
			t.Errorf("seat 1 TurnsBegun %d after its skipped turn, want %d", g.Seats[1].TurnsBegun, begun)
		}
	}
	if g.Seats[1].TurnsBegun != begun+1 {
		t.Errorf("seat 1 TurnsBegun %d after its first real turn, want %d", g.Seats[1].TurnsBegun, begun+1)
	}
	if n := g.TurnSkipsFor(g.Seats[1].ID); n != 0 {
		t.Errorf("seat 1 still owes %d skips", n)
	}
	if len(g.ScopedEffects) != 0 {
		t.Errorf("a spent debt left a record: %+v", g.ScopedEffects)
	}
	if got := countEvents(g, EventTurnSkipped); got != 1 {
		t.Errorf("EventTurnSkipped emitted %d times, want 1", got)
	}
	if got := countEvents(g, EventTurnSkipAdded); got != 1 {
		t.Errorf("EventTurnSkipAdded emitted %d times, want 1", got)
	}
}

// TestASkippedTurnBeginsNothing: no EventTurnBegan for the skipped seat,
// and the event it does emit names the player, the source and a normal
// turn.
func TestASkippedTurnBeginsNothing(t *testing.T) {
	g := newFourPlayerActiveGame(t)
	source := uuid.New()
	skipTurns(t, g, 1, source, 1)
	passTurn(t, g)

	for _, ev := range g.Events {
		if ev.Kind == EventTurnBegan && ev.Actor == g.Seats[1].ID {
			t.Errorf("seat 1's skipped turn began: %+v", ev)
		}
	}
	var skipped *Event
	for i := range g.Events {
		if g.Events[i].Kind == EventTurnSkipped {
			skipped = &g.Events[i]
		}
	}
	if skipped == nil {
		t.Fatal("no EventTurnSkipped")
	}
	if skipped.Actor != g.Seats[1].ID || skipped.Source != source || skipped.Label != "" || skipped.Amount != 0 {
		t.Errorf("EventTurnSkipped = %+v, want seat 1, the source, a normal turn", *skipped)
	}
}

// TestTwoSkipsSkipTwoTurns is CR 614.10a: an Amount of 2, and two records
// of 1, both skip the next two turns. The oldest record is spent first.
func TestTwoSkipsSkipTwoTurns(t *testing.T) {
	t.Run("amount 2", func(t *testing.T) {
		g := newFourPlayerActiveGame(t)
		skipTurns(t, g, 1, uuid.Nil, 2)
		// Seat 1 is skipped in round 1 and in round 2.
		walkTurns(t, g, []turnShape{
			{seat: 2, round: 1}, {seat: 3, round: 1},
			{seat: 0, round: 2}, {seat: 2, round: 2}, {seat: 3, round: 2},
			{seat: 0, round: 3}, {seat: 1, round: 3},
		})
		if got := countEvents(g, EventTurnSkipped); got != 2 {
			t.Errorf("EventTurnSkipped emitted %d times, want 2", got)
		}
	})
	t.Run("two records of 1", func(t *testing.T) {
		g := newFourPlayerActiveGame(t)
		first, second := uuid.New(), uuid.New()
		skipTurns(t, g, 1, first, 1)
		skipTurns(t, g, 1, second, 1)
		if n := g.TurnSkipsFor(g.Seats[1].ID); n != 2 {
			t.Fatalf("TurnSkipsFor = %d, want 2", n)
		}
		walkTurns(t, g, []turnShape{
			{seat: 2, round: 1}, {seat: 3, round: 1},
			{seat: 0, round: 2}, {seat: 2, round: 2}, {seat: 3, round: 2},
			{seat: 0, round: 3}, {seat: 1, round: 3},
		})
		var sources []uuid.UUID
		for _, ev := range g.Events {
			if ev.Kind == EventTurnSkipped {
				sources = append(sources, ev.Source)
			}
		}
		if len(sources) != 2 || sources[0] != first || sources[1] != second {
			t.Errorf("skips spent %v, want the oldest record first: [%v %v]", sources, first, second)
		}
	})
}

// TestASkipThatCoversEveryOtherSeatStillAdvances: every seat owes a skip,
// the walk spends them one pass at a time and ends. The guard never
// trips.
func TestASkipThatCoversEveryOtherSeatStillAdvances(t *testing.T) {
	g := newFourPlayerActiveGame(t)
	for seat := 0; seat < 4; seat++ {
		skipTurns(t, g, seat, uuid.Nil, 1)
	}
	passTurn(t, g)
	// Seats 1, 2, 3 and then 0 itself (the walk wraps to the seat whose
	// turn just ended) each spend their one charge, so the first turn that
	// begins is seat 1's, in round 2: its debt was spent on the way past.
	if got := shapeOf(g); got.seat != 1 || got.round != 2 {
		t.Errorf("after one pass with every seat owing a skip: %+v, want seat 1 in round 2", got)
	}
	if got := countEvents(g, EventTurnSkipped); got != 4 {
		t.Errorf("EventTurnSkipped emitted %d times, want 4", got)
	}
}

// TestASkipCreatedDuringYourOwnTurnDoesNotEndIt is CR 614.10: the turn
// that has begun can no longer be skipped. The cursor stays where it was
// and the skip is taken from the next turn.
func TestASkipCreatedDuringYourOwnTurnDoesNotEndIt(t *testing.T) {
	g := newFourPlayerActiveGame(t)
	skipTurns(t, g, 0, uuid.Nil, 1)
	if got := shapeOf(g); got.seat != 0 {
		t.Fatalf("the turn changed when the skip was created: %+v", got)
	}
	advanceTo(t, g, StepEnd)
	if g.Turn.ActiveSeat != 0 {
		t.Fatalf("seat 0's own turn ended early: seat %d", g.Turn.ActiveSeat)
	}
	// 1, 2, 3, then seat 0's next turn is the one skipped, and seat 1 is up.
	walkTurns(t, g, []turnShape{
		{seat: 1, round: 1}, {seat: 2, round: 1}, {seat: 3, round: 1},
		{seat: 1, round: 2},
	})
}

// TestTheNextTurnSkippedIsTheExtraTurn is the Magosi ruling: a player who
// owes a skip and has an extra turn queued skips THAT turn, because it is
// the next one they would begin, and normal rotation then continues. In
// either order of creation.
func TestTheNextTurnSkippedIsTheExtraTurn(t *testing.T) {
	for name, skipFirst := range map[string]bool{"skip then extra turn": true, "extra turn then skip": false} {
		t.Run(name, func(t *testing.T) {
			g := newFourPlayerActiveGame(t)
			var refs []int
			if skipFirst {
				skipTurns(t, g, 0, uuid.Nil, 1)
				refs = takeExtraTurns(t, g, 0, 1)
			} else {
				refs = takeExtraTurns(t, g, 0, 1)
				skipTurns(t, g, 0, uuid.Nil, 1)
			}
			begun := g.Seats[0].TurnsBegun

			// The extra turn is spent as skipped; the next normal turn is
			// seat 1's, and the debt is gone, so seat 0 plays its next
			// normal turn.
			walkTurns(t, g, []turnShape{
				{seat: 1, round: 1}, {seat: 2, round: 1}, {seat: 3, round: 1},
				{seat: 0, round: 2},
			})
			if g.Seats[0].TurnsBegun != begun+1 {
				t.Errorf("seat 0 TurnsBegun %d, want %d: only the normal turn began", g.Seats[0].TurnsBegun, begun+1)
			}
			var extra *Event
			for i := range g.Events {
				if g.Events[i].Kind == EventTurnSkipped {
					extra = &g.Events[i]
				}
			}
			if extra == nil || extra.Label != "extra" || extra.Amount != refs[0] {
				t.Errorf("EventTurnSkipped = %+v, want the extra turn (ref %d)", extra, refs[0])
			}
			if got := countEvents(g, EventExtraTurnSkipped); got != 0 {
				t.Errorf("EventExtraTurnSkipped emitted %d times: that is the replacement's event, not this one", got)
			}
		})
	}
}

// TestASkippedExtraTurnTakesItsBoundTriggerWithIt: Final Fortune's loss
// is bound to the extra turn's end step. Skipping the turn sweeps it,
// and the game is not lost (CR 614.10a).
func TestASkippedExtraTurnTakesItsBoundTriggerWithIt(t *testing.T) {
	g := newFourPlayerActiveGame(t)
	refs := takeExtraTurns(t, g, 0, 1)
	g.WithWriteLock(func() {
		g.ScheduleDelayedTriggerForEffect(DelayedTrigger{
			Controller: g.Seats[0].ID, Label: "probe — you lose the game", At: StepEnd,
			OnExtraTurn: refs[0],
			Body: testBody(func(g *Game, _ *StackItem) error {
				t.Error("the loss bound to a skipped turn fired")
				return nil
			}),
		})
	})
	skipTurns(t, g, 0, uuid.Nil, 1)

	passTurn(t, g)
	if got := shapeOf(g); got != (turnShape{seat: 1, round: 1}) {
		t.Fatalf("after the skipped extra turn: %+v, want seat 1's normal turn", got)
	}
	if len(g.DelayedTriggers) != 0 {
		t.Errorf("the trigger bound to the skipped turn survived: %d queued", len(g.DelayedTriggers))
	}
	if g.Seats[0].Eliminated {
		t.Error("seat 0 lost the game although its extra turn was skipped")
	}
}

// TestASkippedTurnLeavesTheLastTurnTakenAlone: the per-turn records
// describe the last turn TAKEN (ADR 0188 Decision 1's table). Seat 0
// attacks and casts, seat 1's turn is skipped, and seat 2's turn begins
// with seat 0's record intact and seat 1's untouched.
func TestASkippedTurnLeavesTheLastTurnTakenAlone(t *testing.T) {
	g := newFourPlayerActiveGame(t)
	attacker := uuid.New()
	g.WithWriteLock(func() {
		g.TurnTally.Attacks = []AttackRecord{{Attacker: attacker}}
		g.SpellsCastThisTurn = map[uuid.UUID]CastTally{g.Seats[0].ID: {Total: 2}}
	})
	g.Seats[1].LastTurnAttacks = []AttackRecord{{Attacker: uuid.New()}}
	older := g.Seats[1].LastTurnAttacks[0].Attacker
	skipTurns(t, g, 1, uuid.Nil, 1)

	passTurn(t, g)
	if g.Turn.ActiveSeat != 2 {
		t.Fatalf("active seat %d, want 2", g.Turn.ActiveSeat)
	}
	if got := g.Seats[0].LastTurnAttacks; len(got) != 1 || got[0].Attacker != attacker {
		t.Errorf("seat 0's last turn's attacks = %+v, want the one it declared", got)
	}
	if got := g.Seats[1].LastTurnAttacks; len(got) != 1 || got[0].Attacker != older {
		t.Errorf("seat 1's attacks were overwritten by a turn it did not take: %+v", got)
	}
	if !g.DayNight.PrevTurnKnown || g.DayNight.PrevTurnSpells != 2 {
		t.Errorf("day/night reads %d spells (known=%v), want seat 0's 2 — the last turn taken",
			g.DayNight.PrevTurnSpells, g.DayNight.PrevTurnKnown)
	}
}

// TestAnUpkeepTriggerWaitsPastASkippedTurn: a rebound-shaped trigger,
// "at the beginning of your next upkeep", does not fire in a turn that
// never begins and fires at the next real upkeep (CR 614.10a).
func TestAnUpkeepTriggerWaitsPastASkippedTurn(t *testing.T) {
	g := newFourPlayerActiveGame(t)
	fired := 0
	g.WithWriteLock(func() {
		g.ScheduleDelayedTriggerForEffect(DelayedTrigger{
			Controller: g.Seats[1].ID, Label: "probe — your next upkeep", At: StepUpkeep,
			ControllerTurnOnly: true,
			Body:               testBody(func(*Game, *StackItem) error { fired++; return nil }),
		})
	})
	skipTurns(t, g, 1, uuid.Nil, 1)

	advanceToStepOfSeat(t, g, 2, StepUpkeep)
	settleStack(t, g)
	if fired != 0 || len(g.DelayedTriggers) != 1 {
		t.Fatalf("the trigger fired in a skipped turn or was dropped: fired=%d queued=%d", fired, len(g.DelayedTriggers))
	}
	advanceToStepOfSeat(t, g, 1, StepUpkeep)
	settleStack(t, g)
	if fired != 1 {
		t.Errorf("the trigger fired %d times at seat 1's next real upkeep, want 1", fired)
	}
}

// TestUntilYourNextTurnSurvivesASkippedTurn: "until your next turn"
// waits for the first turn that is not skipped (CR 614.10a), because a
// skipped turn is not counted in TurnsBegun.
func TestUntilYourNextTurnSurvivesASkippedTurn(t *testing.T) {
	g := newFourPlayerActiveGame(t)
	id := pushScopedTestCreature(g, g.Seats[0].ID, 2, 2)
	var d Duration
	g.WithWriteLock(func() { d = g.UntilYourNextTurnDuration(g.Seats[1].ID) })
	registerScopedEffectForTest(t, g, id, []Mod{ModifyPTMod(3, 3)}, d)
	skipTurns(t, g, 1, uuid.Nil, 1)

	passTurn(t, g) // seat 1's turn is skipped, seat 2 begins
	if g.Turn.ActiveSeat != 2 {
		t.Fatalf("active seat %d, want 2", g.Turn.ActiveSeat)
	}
	if c := scopedEffectChar(t, g, id); c.Power != 5 {
		t.Fatalf("the effect ended with a turn that never began: power %d, want 5", c.Power)
	}
	passTurn(t, g) // 3
	passTurn(t, g) // 0
	passTurn(t, g) // 1, a real turn
	if g.Turn.ActiveSeat != 1 {
		t.Fatalf("active seat %d, want 1", g.Turn.ActiveSeat)
	}
	if c := scopedEffectChar(t, g, id); c.Power != 2 {
		t.Errorf("power %d at seat 1's first real turn, want the effect ended (2)", c.Power)
	}
}

// TestADepartedPlayersSkipIsDroppedAndTheirTurnIsPassedAsBefore: a debt
// owed by a player who has left can never be spent (CR 800.4k). The seam
// passes the seat as it always did (counting its turn toward TurnsBegun,
// CR 800.4m) and drops the record, and no skip is announced.
func TestADepartedPlayersSkipIsDroppedAndTheirTurnIsPassedAsBefore(t *testing.T) {
	g := newFourPlayerActiveGame(t)
	skipTurns(t, g, 1, uuid.Nil, 2)
	if err := g.Concede(g.Seats[1].ID); err != nil {
		t.Fatalf("Concede: %v", err)
	}
	if n := g.TurnSkipsFor(g.Seats[1].ID); n != 0 {
		t.Errorf("TurnSkipsFor a player who has left = %d, want 0", n)
	}
	begun := g.Seats[1].TurnsBegun

	passTurn(t, g)
	if got := shapeOf(g); got != (turnShape{seat: 2, round: 1}) {
		t.Fatalf("next turn %+v, want seat 2 (seat 1 has left)", got)
	}
	if g.Seats[1].TurnsBegun != begun+1 {
		t.Errorf("seat 1 TurnsBegun %d, want %d: the passed turn is counted (CR 800.4m)", g.Seats[1].TurnsBegun, begun+1)
	}
	if len(g.ScopedEffects) != 0 {
		t.Errorf("a departed player's debt survived: %+v", g.ScopedEffects)
	}
	if got := countEvents(g, EventTurnSkipped); got != 0 {
		t.Errorf("EventTurnSkipped emitted %d times for a player who had left", got)
	}
}

// TestASkipForAPlayerWhoHasLeftWritesNothing: a skipped turn of a player
// who is not in the game is meaningless (CR 800.4a), as are n < 1 and a
// player who was never seated.
func TestASkipForAPlayerWhoHasLeftWritesNothing(t *testing.T) {
	g := newFourPlayerActiveGame(t)
	if err := g.Concede(g.Seats[1].ID); err != nil {
		t.Fatalf("Concede: %v", err)
	}
	g.WithWriteLock(func() {
		if g.SkipNextTurnsForEffect(g.Seats[1].ID, uuid.Nil, 1, "probe") {
			t.Error("a departed player was given a skip")
		}
		if g.SkipNextTurnsForEffect(g.Seats[2].ID, uuid.Nil, 0, "probe") {
			t.Error("a skip of 0 turns wrote a record")
		}
		if g.SkipNextTurnsForEffect(uuid.New(), uuid.Nil, 1, "probe") {
			t.Error("a player who was never seated was given a skip")
		}
	})
	if len(g.ScopedEffects) != 0 || countEvents(g, EventTurnSkipAdded) != 0 {
		t.Errorf("a refused skip left a trace: %+v", g.ScopedEffects)
	}
}

// TestTurnSkipsRewindWithUndo: the charge is spent copy-on-write, so an
// undo snapshot taken before the rotation keeps it, and taking the turn
// again from the restored game spends it again.
func TestTurnSkipsRewindWithUndo(t *testing.T) {
	g := newFourPlayerActiveGame(t)
	skipTurns(t, g, 1, uuid.Nil, 1)
	before := g.Clone()

	passTurn(t, g)
	if g.Turn.ActiveSeat != 2 || g.TurnSkipsFor(g.Seats[1].ID) != 0 {
		t.Fatalf("setup: seat %d owes %d", g.Turn.ActiveSeat, g.TurnSkipsFor(g.Seats[1].ID))
	}

	g.RestoreFrom(before)
	if g.Turn.ActiveSeat != 0 || g.TurnSkipsFor(g.Seats[1].ID) != 1 {
		t.Fatalf("undo: seat %d owes %d, want seat 0 owing 1", g.Turn.ActiveSeat, g.TurnSkipsFor(g.Seats[1].ID))
	}
	passTurn(t, g)
	if g.Turn.ActiveSeat != 2 || g.TurnSkipsFor(g.Seats[1].ID) != 0 {
		t.Errorf("redo: seat %d owes %d, want seat 2 owing 0", g.Turn.ActiveSeat, g.TurnSkipsFor(g.Seats[1].ID))
	}
	if got := before.TurnSkipsFor(before.Seats[1].ID); got != 1 {
		t.Errorf("the undo snapshot's charge was spent through the live game: %d", got)
	}
}

// TestTurnSkipsSurviveASnapshot: the debt is a record, so a restore point
// carries it and the restored game skips the turn.
func TestTurnSkipsSurviveASnapshot(t *testing.T) {
	g := newFourPlayerActiveGame(t)
	skipTurns(t, g, 1, uuid.Nil, 2)
	passTurn(t, g)
	_, restored := roundTrip(t, g)
	if n := restored.TurnSkipsFor(restored.Seats[1].ID); n != 1 {
		t.Fatalf("restored game owes %d, want the 1 left after one skip", n)
	}
	// Seat 1 was skipped once, so it is seat 2's turn; two more and the
	// debt is spent on the next pass over seat 1.
	passTurn(t, restored)
	passTurn(t, restored)
	passTurn(t, restored)
	if got := shapeOf(restored); got.seat != 2 {
		t.Errorf("restored game: %+v, want seat 2 (seat 1 skipped again)", got)
	}
}

// TestASnapshotNamingSkipTurnsIsRefusedWithoutIt is ADR 0188 Decision 11:
// an older binary has no skipTurns mod kind, and refuses the file instead
// of handing the player a turn their card took.
func TestASnapshotNamingSkipTurnsIsRefusedWithoutIt(t *testing.T) {
	g := newFourPlayerActiveGame(t)
	skipTurns(t, g, 1, uuid.Nil, 1)
	snap := g.CaptureSnapshot()

	spec, ok := modKinds[ModSkipTurns]
	if !ok {
		t.Fatal("skipTurns is not in the mod-kind table")
	}
	delete(modKinds, ModSkipTurns)
	defer func() { modKinds[ModSkipTurns] = spec }()

	if _, err := snap.Restore(); !errors.Is(err, ErrUnknownEffectKey) {
		t.Errorf("Restore without skipTurns = %v, want ErrUnknownEffectKey", err)
	}
	if _, err := snap.RestoreStrict(); !errors.Is(err, ErrUnknownEffectKey) {
		t.Errorf("RestoreStrict without skipTurns = %v, want ErrUnknownEffectKey", err)
	}
}

// TestASkipTurnsRecordWithNoPlayerOrNoTurnsIsRefused: registration
// panics and restore refuses a mod that owes nothing or is owed by no one.
func TestASkipTurnsRecordWithNoPlayerOrNoTurnsIsRefused(t *testing.T) {
	for name, m := range map[string]Mod{
		"no player": {Kind: ModSkipTurns, Amount: 1},
		"no turns":  {Kind: ModSkipTurns, Player: uuid.New()},
	} {
		if skipTurnsModProblem(m) == "" {
			t.Errorf("%s: accepted", name)
		}
	}
	if p := skipTurnsModProblem(Mod{Kind: ModSkipTurns, Player: uuid.New(), Amount: 2}); p != "" {
		t.Errorf("a sound mod refused: %s", p)
	}
	if p := skipTurnsModProblem(Mod{Kind: ModAttackTax}); p != "" {
		t.Errorf("another kind refused: %s", p)
	}

	g := newFourPlayerActiveGame(t)
	skipTurns(t, g, 1, uuid.Nil, 1)
	_, strictErr := func() (*Game, error) {
		snap := g.CaptureSnapshot()
		snap.ScopedEffects[0].Mods[0].Amount = 0
		return snap.RestoreStrict()
	}()
	if !errors.Is(strictErr, ErrUnknownEffectKey) {
		t.Errorf("a restore point owing 0 turns = %v, want ErrUnknownEffectKey", strictErr)
	}

	defer func() {
		if recover() == nil {
			t.Error("registering a skipTurns mod with no player did not panic")
		}
	}()
	g.WithWriteLock(func() {
		g.RegisterScopedRuleEffectForEffect(uuid.Nil, ScopeGame, uuid.Nil,
			[]Mod{{Kind: ModSkipTurns, Amount: 1}}, IndefiniteDuration(), "probe")
	})
}
