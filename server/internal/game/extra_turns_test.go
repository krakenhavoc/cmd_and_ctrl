package game

import (
	"testing"

	"github.com/google/uuid"
)

// extra_turns_test.go pins CR 500.7 extra turns (ADR 0059 Decision 5,
// #753): the queue is a stack taken most-recent-first, normal rotation
// resumes from the seat the extra turns interrupted, an extra turn is a
// full turn for every per-turn reset, and a queued turn of a player who
// has left the game doesn't begin (CR 800.4k).

// takeExtraTurns queues n extra turns for `seat` under the write lock.
func takeExtraTurns(t *testing.T, g *Game, seat, n int) []int {
	t.Helper()
	var refs []int
	g.WithWriteLock(func() {
		refs = g.TakeExtraTurnsForEffect(g.Seats[seat].ID, uuid.Nil, n)
	})
	if len(refs) != n {
		t.Fatalf("TakeExtraTurnsForEffect(seat %d, %d) = %v", seat, n, refs)
	}
	return refs
}

// passTurn ends the current turn through the sandbox verb, which goes
// through the same rotation seam as the cursor walking past cleanup.
func passTurn(t *testing.T, g *Game) {
	t.Helper()
	if err := g.EndTurnNowForTest(); err != nil {
		t.Fatalf("PassTurn: %v", err)
	}
}

type turnShape struct {
	seat  int
	extra bool
	round int
}

func shapeOf(g *Game) turnShape {
	return turnShape{seat: g.Turn.ActiveSeat, extra: g.Turn.Extra, round: g.Turn.Round}
}

// TestExtraTurnGoesDirectlyAfterThisOneAndRotationResumes is ADR 0059
// test 2: seat 0 gives seat 2 an extra turn in a four-player game. Seat
// 2's extra turn follows seat 0's, then seat 1 takes its normal turn —
// rotation resumes from the seat whose normal turn was interrupted —
// and Round moves only on the wrap.
func TestExtraTurnGoesDirectlyAfterThisOneAndRotationResumes(t *testing.T) {
	g := newFourPlayerActiveGame(t)
	if g.Turn.ActiveSeat != 0 {
		t.Fatalf("setup: seat %d starts, want 0", g.Turn.ActiveSeat)
	}
	refs := takeExtraTurns(t, g, 2, 1)
	begun := g.Seats[2].TurnsBegun
	seq := g.Turn.Seq

	want := []turnShape{
		{seat: 2, extra: true, round: 1},
		{seat: 1, round: 1},
		{seat: 2, round: 1},
		{seat: 3, round: 1},
		{seat: 0, round: 2},
	}
	for i, w := range want {
		passTurn(t, g)
		if got := shapeOf(g); got != w {
			t.Fatalf("turn %d after the extra-turn effect: %+v, want %+v", i+1, got, w)
		}
		if g.Turn.Seq != seq+i+1 {
			t.Errorf("turn %d: Seq %d, want %d (one per turn, extra turns included)", i+1, g.Turn.Seq, seq+i+1)
		}
		if i == 0 {
			if g.Turn.ExtraRef != refs[0] {
				t.Errorf("extra turn's ExtraRef %d, want %d", g.Turn.ExtraRef, refs[0])
			}
			if g.Turn.OrderSeat != 0 {
				t.Errorf("OrderSeat %d during the extra turn, want 0 (seat 0's normal turn was interrupted)", g.Turn.OrderSeat)
			}
			if g.Seats[2].TurnsBegun != begun+1 {
				t.Errorf("seat 2 TurnsBegun %d, want %d: an extra turn is that seat's turn", g.Seats[2].TurnsBegun, begun+1)
			}
		}
		if i > 0 && (g.Turn.Extra || g.Turn.ExtraRef != 0) {
			t.Errorf("turn %d is a normal turn but reads extra=%v ref=%d", i+1, g.Turn.Extra, g.Turn.ExtraRef)
		}
	}
	if len(g.ExtraTurns) != 0 {
		t.Errorf("queue not drained: %+v", g.ExtraTurns)
	}
}

// TestExtraTurnsTakeTheMostRecentFirst is ADR 0059 test 3 (CR 500.7):
// Time Stretch gives seat 1 two turns; during the first of them a Time
// Warp gives seat 0 one. The Warp turn comes before the second Stretch
// turn, and normal rotation then resumes after the ORIGINAL seat.
func TestExtraTurnsTakeTheMostRecentFirst(t *testing.T) {
	g := newFourPlayerActiveGame(t)
	stretch := takeExtraTurns(t, g, 1, 2)
	if stretch[0] >= stretch[1] {
		t.Fatalf("refs %v should come out in the order the turns are taken", stretch)
	}
	if q := g.ExtraTurnsQueued(); len(q) != 2 || q[0].Ref != stretch[0] || q[1].Ref != stretch[1] {
		t.Fatalf("ExtraTurnsQueued = %+v, want refs %v next-first", q, stretch)
	}

	passTurn(t, g)
	if g.Turn.ActiveSeat != 1 || g.Turn.ExtraRef != stretch[0] {
		t.Fatalf("first Stretch turn: seat %d ref %d", g.Turn.ActiveSeat, g.Turn.ExtraRef)
	}
	warp := takeExtraTurns(t, g, 0, 1)

	passTurn(t, g)
	if g.Turn.ActiveSeat != 0 || g.Turn.ExtraRef != warp[0] {
		t.Fatalf("after the first Stretch turn: seat %d ref %d, want the Warp turn (seat 0, ref %d)",
			g.Turn.ActiveSeat, g.Turn.ExtraRef, warp[0])
	}
	passTurn(t, g)
	if g.Turn.ActiveSeat != 1 || g.Turn.ExtraRef != stretch[1] {
		t.Fatalf("after the Warp turn: seat %d ref %d, want the second Stretch turn", g.Turn.ActiveSeat, g.Turn.ExtraRef)
	}
	passTurn(t, g)
	if got := shapeOf(g); got != (turnShape{seat: 1, round: 1}) {
		t.Fatalf("after the extra turns: %+v, want seat 1's normal turn in round 1", got)
	}
}

// TestExtraTurnResetsPerTurnStateForTheSameSeat is ADR 0059 test 1: a
// seat taking two turns in a row gets every per-turn reset on the
// second — the land drop, the planeswalker's loyalty activation, the
// once-per-turn gates — and "until end of turn" effects end between
// them.
func TestExtraTurnResetsPerTurnStateForTheSameSeat(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	takeExtraTurns(t, g, me.Seat, 1)
	gate := TallyKey(uuid.New(), "once each turn")
	walker := uuid.New()
	g.WithWriteLock(func() {
		g.LandsPlayedThisTurn = map[uuid.UUID]int{me.ID: 1}
		g.LoyaltyActivatedThisTurn = map[uuid.UUID]bool{walker: true}
		if g.TurnTally.Triggered == nil {
			g.TurnTally.Triggered = map[string]int{}
		}
		g.TurnTally.Triggered[gate] = 1
	})
	layers := g.layerVersion.Load()
	begun := me.TurnsBegun
	round := g.Turn.Round

	passTurn(t, g)
	if g.Turn.ActiveSeat != me.Seat || !g.Turn.Extra {
		t.Fatalf("seat %d extra=%v, want %s's extra turn", g.Turn.ActiveSeat, g.Turn.Extra, me.Name)
	}
	if g.Turn.Round != round {
		t.Errorf("Round %d, want %d: an extra turn never moves the round", g.Turn.Round, round)
	}
	if me.TurnsBegun != begun+1 {
		t.Errorf("TurnsBegun %d, want %d", me.TurnsBegun, begun+1)
	}
	if len(g.LandsPlayedThisTurn) != 0 || g.LoyaltyActivatedThisTurn[walker] || g.TurnTally.Triggered[gate] != 0 {
		t.Errorf("per-turn state survived into the extra turn: lands %v loyalty %v gate %d",
			g.LandsPlayedThisTurn, g.LoyaltyActivatedThisTurn, g.TurnTally.Triggered[gate])
	}
	if g.layerVersion.Load() == layers {
		t.Error("layerVersion did not move as the extra turn began")
	}
	if g.Turn.Step != StepUpkeep {
		t.Errorf("the extra turn should have untapped and landed on upkeep, got %s", g.Turn.Step)
	}

	passTurn(t, g)
	if g.Turn.ActiveSeat == me.Seat || g.Turn.Extra {
		t.Fatalf("after the extra turn: seat %d extra=%v, want the opponent's normal turn", g.Turn.ActiveSeat, g.Turn.Extra)
	}
}

// TestExtraTurnOfADepartedPlayerDoesNotBegin is ADR 0059 test 4
// (CR 800.4k): a queued extra turn of a player who concedes is dropped
// as it would begin, still counts toward their TurnsBegun (CR 800.4m),
// and a delayed trigger bound to it is swept — Final Fortune's ruling,
// "If you end up skipping the extra turn … you do not lose the game."
func TestExtraTurnOfADepartedPlayerDoesNotBegin(t *testing.T) {
	g := newFourPlayerActiveGame(t)
	refs := takeExtraTurns(t, g, 2, 1)
	fired := 0
	g.WithWriteLock(func() {
		if g.ScheduleDelayedTriggerForEffect(DelayedTrigger{
			Controller: g.Seats[2].ID, Label: "probe — that turn's end step", At: StepEnd,
			OnExtraTurn: refs[0],
			Body:        testBody(func(*Game, *StackItem) error { fired++; return nil }),
		}) == uuid.Nil {
			t.Fatal("a trigger bound to a queued extra turn was refused")
		}
	})
	if err := g.Concede(g.Seats[2].ID); err != nil {
		t.Fatalf("Concede: %v", err)
	}
	begun := g.Seats[2].TurnsBegun

	passTurn(t, g)
	if got := shapeOf(g); got != (turnShape{seat: 1, round: 1}) {
		t.Fatalf("next turn %+v, want seat 1's normal turn (seat 2 has left)", got)
	}
	if g.Seats[2].TurnsBegun != begun+1 {
		t.Errorf("seat 2 TurnsBegun %d, want %d: the dropped turn would have begun (CR 800.4m)", g.Seats[2].TurnsBegun, begun+1)
	}
	if len(g.DelayedTriggers) != 0 {
		t.Errorf("the trigger bound to the dropped turn survived: %d queued", len(g.DelayedTriggers))
	}
	if len(g.ExtraTurns) != 0 {
		t.Errorf("queue not drained: %+v", g.ExtraTurns)
	}
	if fired != 0 {
		t.Errorf("bound trigger fired %d times", fired)
	}

	g.WithWriteLock(func() {
		if refs := g.TakeExtraTurnsForEffect(g.Seats[2].ID, uuid.Nil, 1); refs != nil {
			t.Errorf("a departed player was given an extra turn: %v", refs)
		}
	})
}

// TestDelayedTriggerBoundToAnExtraTurnFiresOnlyInIt is ADR 0059
// test 11: "at the beginning of that turn's end step" does not fire in
// the end step of the turn the effect resolved in, fires in the extra
// turn's, and a bound trigger for a turn no longer pending is refused.
func TestDelayedTriggerBoundToAnExtraTurnFiresOnlyInIt(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	refs := takeExtraTurns(t, g, me.Seat, 1)
	fired := 0
	g.WithWriteLock(func() {
		g.ScheduleDelayedTriggerForEffect(DelayedTrigger{
			Controller: me.ID, Label: "probe — that turn's end step", At: StepEnd,
			OnExtraTurn: refs[0],
			Body:        testBody(func(*Game, *StackItem) error { fired++; return nil }),
		})
	})

	advanceTo(t, g, StepEnd)
	settleStack(t, g)
	if fired != 0 || len(g.DelayedTriggers) != 1 {
		t.Fatalf("fired in this turn's end step: fired=%d queued=%d", fired, len(g.DelayedTriggers))
	}

	passTurn(t, g)
	if !g.Turn.Extra || g.Turn.ExtraRef != refs[0] {
		t.Fatalf("expected the extra turn, got %+v", g.Turn)
	}
	advanceTo(t, g, StepEnd)
	settleStack(t, g)
	if fired != 1 {
		t.Fatalf("bound trigger fired %d times in its own turn's end step, want 1", fired)
	}

	g.WithWriteLock(func() {
		if id := g.ScheduleDelayedTriggerForEffect(DelayedTrigger{
			Controller: me.ID, Label: "probe — stale binding", At: StepEnd, OnExtraTurn: refs[0] + 99,
			Body: testBody(func(*Game, *StackItem) error { return nil }),
		}); id != uuid.Nil {
			t.Error("a trigger bound to a turn that is neither running nor queued was queued")
		}
	})
}

// TestBoundTriggerIsSweptWhenItsTurnEndsEarly: the extra turn ends
// without reaching its end step (the sandbox pass_turn verb here; a
// departure in play), so the trigger bound to that end step can never
// fire and is dropped at the turn change.
func TestBoundTriggerIsSweptWhenItsTurnEndsEarly(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	refs := takeExtraTurns(t, g, me.Seat, 1)
	g.WithWriteLock(func() {
		g.ScheduleDelayedTriggerForEffect(DelayedTrigger{
			Controller: me.ID, Label: "probe", At: StepEnd, OnExtraTurn: refs[0],
			Body: testBody(func(*Game, *StackItem) error { return nil }),
		})
	})
	passTurn(t, g) // into the extra turn
	if len(g.DelayedTriggers) != 1 {
		t.Fatalf("swept at the start of its own turn: %d queued", len(g.DelayedTriggers))
	}
	passTurn(t, g) // out of it, never having reached its end step
	if len(g.DelayedTriggers) != 0 {
		t.Errorf("trigger for an extra turn that has ended survived: %d queued", len(g.DelayedTriggers))
	}
}

// TestExtraTurnQueueRewindsWithUndo: the queue is game state. An undo
// past the effect that queued a turn un-queues it, and one taken after
// restores it (ADR 0059 test 18, engine half).
func TestExtraTurnQueueRewindsWithUndo(t *testing.T) {
	g := newActiveGame(t)
	before := g.Clone()
	takeExtraTurns(t, g, 1, 2)
	after := g.Clone()

	g.RestoreFrom(before)
	if len(g.ExtraTurns) != 0 || g.NextExtraRef != 0 {
		t.Fatalf("undo left the queue: %+v next=%d", g.ExtraTurns, g.NextExtraRef)
	}
	g.RestoreFrom(after)
	if len(g.ExtraTurns) != 2 || g.NextExtraRef != 2 {
		t.Fatalf("redo lost the queue: %+v next=%d", g.ExtraTurns, g.NextExtraRef)
	}
	// Taking a turn off the restored queue must not rewrite the clone
	// it came from.
	passTurn(t, g)
	takeExtraTurns(t, g, 0, 1)
	if len(after.ExtraTurns) != 2 || after.ExtraTurns[0].Seat != 1 || after.ExtraTurns[1].Seat != 1 {
		t.Errorf("the undo snapshot's queue was written through: %+v", after.ExtraTurns)
	}
}

// TestExtraTurnSurvivesASnapshot: a queued turn and a bound trigger
// come back from a restore point, and the restored game takes the turn.
func TestExtraTurnSurvivesASnapshot(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	refs := takeExtraTurns(t, g, me.Seat, 1)
	g.WithWriteLock(func() {
		g.ScheduleDelayedTriggerForEffect(DelayedTrigger{
			Controller: me.ID, Label: "probe", At: StepEnd, OnExtraTurn: refs[0],
			Body: carriedTestBodyKey,
		})
	})
	_, restored := roundTrip(t, g)
	if len(restored.ExtraTurns) != 1 || restored.ExtraTurns[0].Ref != refs[0] || restored.NextExtraRef != refs[0] {
		t.Fatalf("queue after restore: %+v next=%d", restored.ExtraTurns, restored.NextExtraRef)
	}
	if len(restored.DelayedTriggers) != 1 || restored.DelayedTriggers[0].OnExtraTurn != refs[0] {
		t.Fatalf("bound trigger after restore: %+v", restored.DelayedTriggers)
	}
	passTurn(t, restored)
	if restored.Turn.ActiveSeat != me.Seat || !restored.Turn.Extra {
		t.Fatalf("restored game did not take the queued turn: %+v", restored.Turn)
	}
}
