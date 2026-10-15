package game

import (
	"errors"
	"testing"

	"github.com/google/uuid"
)

// turn_skip_order_test.go pins ADR 0188 Decision 12 (#3561): when a
// player owes "skip your next turn" and an extra-turn skip applies to
// the same extra turn, the turn change pauses and asks which skip takes
// it (CR 616.1). Each test drives the real rotation seam.

// skipExtraTurnsOf is a Trouble in Pairs-shaped replacement: a pure
// cancel of `seat`'s extra turns.
func skipExtraTurnsOf(seat int) ReplacementEffect {
	return ReplacementEffect{
		Watches: []EventKind{EventExtraTurnBegin},
		AppliesTo: func(ev *ReplacementEvent, _ *Game, _ *Card) bool {
			return ev.Kind == RepEventExtraTurn && ev.ExtraTurnSeat == seat
		},
		Replace: func(ev *ReplacementEvent, _ *Game, _ *Card) error {
			ev.Cancel()
			return nil
		},
		PureCancel: true,
		Label:      "probe — skip that extra turn",
	}
}

// pausedTable is the acceptance board on four seats: seat 0 owes a skip,
// has `extras` extra turns queued, and a replacement skips seat 0's
// extra turns. The returned game has not ended its turn yet.
func pausedTable(t *testing.T, extras int) (*Game, []int) {
	t.Helper()
	g := newFourPlayerActiveGame(t)
	g.WithWriteLock(func() { g.RegisterReplacementForTest(skipExtraTurnsOf(0)) })
	skipTurns(t, g, 0, uuid.Nil, 1)
	refs := takeExtraTurns(t, g, 0, extras)
	return g, refs
}

// theQuestion is the open turn_skip_order prompt, or a test failure.
func theQuestion(t *testing.T, g *Game) *PendingChoice {
	t.Helper()
	for _, c := range g.PendingChoices {
		if c.Kind == PendingChoiceTurnSkipOrder {
			return c
		}
	}
	t.Fatalf("no turn_skip_order prompt open (%d prompts, pending %v)", len(g.PendingChoices), g.TurnChangePending)
	return nil
}

func answer(t *testing.T, g *Game, useSkip bool) {
	t.Helper()
	c := theQuestion(t, g)
	if err := g.ResolveTurnSkipOrder(c.ID, c.Chooser, useSkip); err != nil {
		t.Fatalf("ResolveTurnSkipOrder(%v): %v", useSkip, err)
	}
}

// TestBothSkipsPauseTheTurnChange: the question is queued and nothing
// begins — Seq unchanged, no EventTurnBegan, nobody holds priority, the
// flag is set and the extra turn is back on top of the queue.
func TestBothSkipsPauseTheTurnChange(t *testing.T) {
	g, refs := pausedTable(t, 1)
	seq := g.Turn.Seq
	began := countEvents(g, EventTurnBegan)
	steps := countEvents(g, EventStepBegan)

	passTurn(t, g)
	c := theQuestion(t, g)
	if got := countEvents(g, EventStepBegan); got != steps {
		t.Errorf("%d steps began while paused", got-steps)
	}
	if !g.TurnChangePending || g.Turn.Seq != seq || g.Turn.ActiveSeat != 0 || g.Turn.PriorityHolder != NoPriority {
		t.Fatalf("not paused: pending %v seq %d (was %d) seat %d priority %d",
			g.TurnChangePending, g.Turn.Seq, seq, g.Turn.ActiveSeat, g.Turn.PriorityHolder)
	}
	if got := countEvents(g, EventTurnBegan); got != began {
		t.Errorf("%d turns began while paused", got-began)
	}
	if c.Chooser != g.Seats[0].ID || c.TurnSkipOrder == nil || c.TurnSkipOrder.ExtraTurnRef != refs[0] {
		t.Errorf("prompt %+v / %+v, want seat 0 asked about ref %d", c, c.TurnSkipOrder, refs[0])
	}
	if c.AutoAnswerKey != TurnSkipOrderAutoAnswerKey || c.midResolution {
		t.Errorf("key %q midResolution %v", c.AutoAnswerKey, c.midResolution)
	}
	if q := g.ExtraTurnsQueued(); len(q) != 1 || q[0].Ref != refs[0] {
		t.Errorf("queue %+v, want the questioned turn on top", q)
	}
}

// TestUseMySkipSpendsTheChargeAndTheNextTurnBegins is answer 0: the
// charge is spent, no EventExtraTurnSkipped, and seat 1's turn begins
// with its step hooks (the untap step walks on to upkeep).
func TestUseMySkipSpendsTheChargeAndTheNextTurnBegins(t *testing.T) {
	g, refs := pausedTable(t, 1)
	seq := g.Turn.Seq
	passTurn(t, g)
	answer(t, g, true)

	if g.TurnChangePending || len(g.PendingChoices) != 0 {
		t.Fatalf("still paused: pending %v, %d prompts", g.TurnChangePending, len(g.PendingChoices))
	}
	if got := shapeOf(g); got != (turnShape{seat: 1, round: 1}) || g.Turn.Seq != seq+1 {
		t.Fatalf("after answer 0: %+v seq %d, want seat 1's normal turn at seq %d", got, g.Turn.Seq, seq+1)
	}
	if g.Turn.Step != StepUpkeep {
		t.Errorf("step %s, want upkeep: the new turn's entry hooks did not run", g.Turn.Step)
	}
	if got := g.TurnSkipsFor(g.Seats[0].ID); got != 0 {
		t.Errorf("seat 0 owes %d, want 0", got)
	}
	if got := countEvents(g, EventExtraTurnSkipped); got != 0 {
		t.Errorf("EventExtraTurnSkipped %d times, want 0", got)
	}
	var skipped *Event
	for i := range g.Events {
		if g.Events[i].Kind == EventTurnSkipped {
			skipped = &g.Events[i]
		}
	}
	if skipped == nil || skipped.Amount != refs[0] || skipped.Label != "extra" {
		t.Errorf("EventTurnSkipped = %+v, want the extra turn %d", skipped, refs[0])
	}
}

// TestLetTheOtherSkipItKeepsTheCharge is answer 1: EventExtraTurnSkipped,
// and the charge is still owed — so seat 0's next normal turn is skipped.
func TestLetTheOtherSkipItKeepsTheCharge(t *testing.T) {
	g, _ := pausedTable(t, 1)
	passTurn(t, g)
	answer(t, g, false)

	if got := shapeOf(g); got != (turnShape{seat: 1, round: 1}) {
		t.Fatalf("after answer 1: %+v, want seat 1", got)
	}
	if got := countEvents(g, EventExtraTurnSkipped); got != 1 {
		t.Errorf("EventExtraTurnSkipped %d times, want 1", got)
	}
	if got := g.TurnSkipsFor(g.Seats[0].ID); got != 1 {
		t.Errorf("seat 0 owes %d, want 1", got)
	}
	walkTurns(t, g, []turnShape{{seat: 2, round: 1}, {seat: 3, round: 1}, {seat: 1, round: 2}})
}

// TestOnlyOneSkipNeverAsks: with only one of the two applying there is
// no question, in both directions.
func TestOnlyOneSkipNeverAsks(t *testing.T) {
	t.Run("only the debt", func(t *testing.T) {
		g := newFourPlayerActiveGame(t)
		skipTurns(t, g, 0, uuid.Nil, 1)
		takeExtraTurns(t, g, 0, 1)
		passTurn(t, g)
		if g.TurnChangePending || len(g.PendingChoices) != 0 {
			t.Fatalf("asked: pending %v, %d prompts", g.TurnChangePending, len(g.PendingChoices))
		}
		if got := shapeOf(g); got != (turnShape{seat: 1, round: 1}) || g.TurnSkipsFor(g.Seats[0].ID) != 0 {
			t.Errorf("%+v owing %d, want seat 1 and the debt spent", got, g.TurnSkipsFor(g.Seats[0].ID))
		}
	})
	t.Run("only the extra-turn skip", func(t *testing.T) {
		g := newFourPlayerActiveGame(t)
		g.WithWriteLock(func() { g.RegisterReplacementForTest(skipExtraTurnsOf(0)) })
		takeExtraTurns(t, g, 0, 1)
		passTurn(t, g)
		if g.TurnChangePending || len(g.PendingChoices) != 0 {
			t.Fatalf("asked: pending %v, %d prompts", g.TurnChangePending, len(g.PendingChoices))
		}
		if got := countEvents(g, EventExtraTurnSkipped); got != 1 {
			t.Errorf("EventExtraTurnSkipped %d times, want 1", got)
		}
	})
}

// TestTwoExtraTurnsAskInTurnAndSpendOneDebtOnce: two queued extra turns
// of the same player, both meeting the condition. Answer 1 keeps the
// debt, so the second turn asks again; answer 0 there spends it. And
// answer 0 on the first leaves no debt, so the second is not asked about.
func TestTwoExtraTurnsAskInTurnAndSpendOneDebtOnce(t *testing.T) {
	t.Run("let, then use", func(t *testing.T) {
		g, refs := pausedTable(t, 2)
		passTurn(t, g)
		if c := theQuestion(t, g); c.TurnSkipOrder.ExtraTurnRef != refs[0] {
			t.Fatalf("first question about ref %d, want %d", c.TurnSkipOrder.ExtraTurnRef, refs[0])
		}
		answer(t, g, false)
		c := theQuestion(t, g)
		if c.TurnSkipOrder.ExtraTurnRef != refs[1] || !g.TurnChangePending {
			t.Fatalf("second question about ref %d (pending %v), want %d", c.TurnSkipOrder.ExtraTurnRef, g.TurnChangePending, refs[1])
		}
		answer(t, g, true)
		if got := shapeOf(g); got != (turnShape{seat: 1, round: 1}) {
			t.Fatalf("%+v, want seat 1", got)
		}
		if g.TurnSkipsFor(g.Seats[0].ID) != 0 || countEvents(g, EventTurnSkipped) != 1 || countEvents(g, EventExtraTurnSkipped) != 1 {
			t.Errorf("owed %d, turn_skipped %d, extra_turn_skipped %d; want 0, 1, 1",
				g.TurnSkipsFor(g.Seats[0].ID), countEvents(g, EventTurnSkipped), countEvents(g, EventExtraTurnSkipped))
		}
	})
	t.Run("use, then no question", func(t *testing.T) {
		g, _ := pausedTable(t, 2)
		passTurn(t, g)
		answer(t, g, true)
		if g.TurnChangePending || len(g.PendingChoices) != 0 {
			t.Fatalf("asked again with the debt spent: %d prompts", len(g.PendingChoices))
		}
		if g.TurnSkipsFor(g.Seats[0].ID) != 0 || countEvents(g, EventTurnSkipped) != 1 || countEvents(g, EventExtraTurnSkipped) != 1 {
			t.Errorf("owed %d, turn_skipped %d, extra_turn_skipped %d; want 0, 1, 1",
				g.TurnSkipsFor(g.Seats[0].ID), countEvents(g, EventTurnSkipped), countEvents(g, EventExtraTurnSkipped))
		}
	})
}

// TestTheCursorVerbsAreRefusedWhileAsked: pass_turn, advance_step,
// pass_priority and end_turn are refused with ErrTurnChangePending.
func TestTheCursorVerbsAreRefusedWhileAsked(t *testing.T) {
	g, _ := pausedTable(t, 1)
	passTurn(t, g)
	if err := g.PassTurn(); !errors.Is(err, ErrTurnChangePending) {
		t.Errorf("PassTurn = %v", err)
	}
	if _, err := g.AdvanceStep(); !errors.Is(err, ErrTurnChangePending) {
		t.Errorf("AdvanceStep = %v", err)
	}
	if err := g.PassPriority(); !errors.Is(err, ErrTurnChangePending) && !errors.Is(err, ErrNoPriority) {
		t.Errorf("PassPriority = %v", err)
	}
	if err := g.EndTurnByPassing(); !errors.Is(err, ErrTurnChangePending) {
		t.Errorf("EndTurnByPassing = %v", err)
	}
	if !g.TurnChangePending || g.Turn.ActiveSeat != 0 {
		t.Errorf("a refused verb moved the table: pending %v seat %d", g.TurnChangePending, g.Turn.ActiveSeat)
	}
}

// TestAnotherSeatConcedingDoesNotRotateTwice: the question stays open
// and the seam does not run again; answering then moves on once.
func TestAnotherSeatConcedingDoesNotRotateTwice(t *testing.T) {
	g, _ := pausedTable(t, 1)
	passTurn(t, g)
	seq := g.Turn.Seq
	if err := g.Concede(g.Seats[2].ID); err != nil {
		t.Fatalf("Concede: %v", err)
	}
	if !g.TurnChangePending || g.Turn.Seq != seq {
		t.Fatalf("a bystander's concession moved the turn: pending %v seq %d (was %d)", g.TurnChangePending, g.Turn.Seq, seq)
	}
	theQuestion(t, g)
	answer(t, g, true)
	if got := shapeOf(g); got != (turnShape{seat: 1, round: 1}) || g.Turn.Seq != seq+1 {
		t.Errorf("%+v seq %d, want seat 1 at seq %d", got, g.Turn.Seq, seq+1)
	}
}

// TestTheOldActivePlayerConcedingDoesNotRotateTwice: seat 0's turn has
// ended and seat 1, who owes a skip, is asked about the extra turn seat
// 0 gave them. Seat 0 — still the active seat on the cursor — concedes.
// Its turn is already over, so the seam must not end it again (which
// would ask a second question); the open one still decides.
func TestTheOldActivePlayerConcedingDoesNotRotateTwice(t *testing.T) {
	g := newFourPlayerActiveGame(t)
	g.WithWriteLock(func() { g.RegisterReplacementForTest(skipExtraTurnsOf(1)) })
	skipTurns(t, g, 1, uuid.Nil, 1)
	takeExtraTurns(t, g, 1, 1)
	passTurn(t, g)
	if c := theQuestion(t, g); c.Chooser != g.Seats[1].ID || g.Turn.ActiveSeat != 0 {
		t.Fatalf("setup: chooser %v, cursor seat %d", c.Chooser, g.Turn.ActiveSeat)
	}
	seq := g.Turn.Seq
	if err := g.Concede(g.Seats[0].ID); err != nil {
		t.Fatalf("Concede: %v", err)
	}
	if len(g.PendingChoices) != 1 || !g.TurnChangePending || g.Turn.Seq != seq {
		t.Fatalf("the old active seat's concession rotated again: %d prompts, pending %v, seq %d (was %d)",
			len(g.PendingChoices), g.TurnChangePending, g.Turn.Seq, seq)
	}
	answer(t, g, true)
	if got := shapeOf(g); got.seat != 1 || got.extra || g.Turn.Seq != seq+1 {
		t.Errorf("%+v seq %d, want seat 1's normal turn at seq %d", got, g.Turn.Seq, seq+1)
	}
}

// TestTheCursorWalkingPastCleanupPausesOnce: the ordinary way a turn
// ends — advance_step through the cleanup step — pauses exactly once,
// and no step of the next turn is entered while it waits.
func TestTheCursorWalkingPastCleanupPausesOnce(t *testing.T) {
	g, _ := pausedTable(t, 1)
	seq := g.Turn.Seq
	for i := 0; i < 40 && !g.TurnChangePending; i++ {
		if len(g.PendingChoices) > 0 {
			t.Fatalf("a prompt other than the question stopped the walk: %s", g.PendingChoices[0].Kind)
		}
		if _, err := g.AdvanceStep(); err != nil {
			t.Fatalf("AdvanceStep %d: %v", i, err)
		}
	}
	if !g.TurnChangePending || len(g.PendingChoices) != 1 {
		t.Fatalf("pending %v with %d prompts, want one question", g.TurnChangePending, len(g.PendingChoices))
	}
	if g.Turn.Seq != seq || g.Turn.ActiveSeat != 0 || g.Turn.Step != StepCleanup {
		t.Errorf("cursor %+v, want seat 0's cleanup at seq %d", g.Turn, seq)
	}
	answer(t, g, true)
	if got := shapeOf(g); got != (turnShape{seat: 1, round: 1}) || g.Turn.Step != StepUpkeep {
		t.Errorf("%+v at %s, want seat 1's upkeep", got, g.Turn.Step)
	}
}

// TestTheChoosersConcessionDropsTheTurnAndRotationContinues: the
// departed player's extra turn doesn't begin (CR 800.4k), and the next
// seat's turn does.
func TestTheChoosersConcessionDropsTheTurnAndRotationContinues(t *testing.T) {
	g, _ := pausedTable(t, 1)
	passTurn(t, g)
	seq := g.Turn.Seq
	if err := g.Concede(g.Seats[0].ID); err != nil {
		t.Fatalf("Concede: %v", err)
	}
	if g.TurnChangePending || len(g.PendingChoices) != 0 {
		t.Fatalf("still paused after the chooser left: pending %v, %d prompts", g.TurnChangePending, len(g.PendingChoices))
	}
	if got := shapeOf(g); got != (turnShape{seat: 1, round: 1}) || g.Turn.Seq != seq+1 {
		t.Errorf("%+v seq %d, want seat 1 at seq %d", got, g.Turn.Seq, seq+1)
	}
	if len(g.ExtraTurns) != 0 {
		t.Errorf("the departed seat's extra turn is still queued: %+v", g.ExtraTurns)
	}
	if g.Turn.Step != StepUpkeep {
		t.Errorf("step %s, want upkeep", g.Turn.Step)
	}
}

// TestUndoAcrossTheQuestion: undoing the turn-ending commit leaves no
// pause; undoing the answer reopens the question.
func TestUndoAcrossTheQuestion(t *testing.T) {
	g, _ := pausedTable(t, 1)
	beforeEnd := g.Clone()
	passTurn(t, g)
	beforeAnswer := g.Clone()
	answer(t, g, true)

	g.RestoreFrom(beforeAnswer)
	if !g.TurnChangePending || g.TurnSkipsFor(g.Seats[0].ID) != 1 {
		t.Fatalf("undo of the answer: pending %v owing %d", g.TurnChangePending, g.TurnSkipsFor(g.Seats[0].ID))
	}
	answer(t, g, false)
	if g.TurnSkipsFor(g.Seats[0].ID) != 1 || g.Turn.ActiveSeat != 1 {
		t.Errorf("the reopened question answered 1: owing %d seat %d", g.TurnSkipsFor(g.Seats[0].ID), g.Turn.ActiveSeat)
	}

	g.RestoreFrom(beforeEnd)
	if g.TurnChangePending || len(g.PendingChoices) != 0 || g.Turn.ActiveSeat != 0 {
		t.Errorf("undo of the turn end: pending %v, %d prompts, seat %d", g.TurnChangePending, len(g.PendingChoices), g.Turn.ActiveSeat)
	}
}

// TestARestorePointTakenMidQuestionRestoresMidQuestion, and an older
// binary — one whose pending-choice table lacks the kind — refuses it.
func TestARestorePointTakenMidQuestionRestoresMidQuestion(t *testing.T) {
	g, refs := pausedTable(t, 1)
	passTurn(t, g)
	_, restored := roundTrip(t, g)
	if !restored.TurnChangePending {
		t.Fatal("the restored game is not between turns")
	}
	c := theQuestion(t, restored)
	if c.TurnSkipOrder == nil || c.TurnSkipOrder.ExtraTurnRef != refs[0] || c.AutoAnswerKey != TurnSkipOrderAutoAnswerKey {
		t.Fatalf("restored prompt %+v / %+v", c, c.TurnSkipOrder)
	}
	restored.WithWriteLock(func() { restored.RegisterReplacementForTest(skipExtraTurnsOf(0)) })
	answer(t, restored, false)
	if restored.Turn.ActiveSeat != 1 || restored.TurnSkipsFor(restored.Seats[0].ID) != 1 {
		t.Errorf("restored answer 1: seat %d owing %d", restored.Turn.ActiveSeat, restored.TurnSkipsFor(restored.Seats[0].ID))
	}

	snap := g.CaptureSnapshot()
	blocks := choiceGateDecisions[PendingChoiceTurnSkipOrder]
	delete(choiceGateDecisions, PendingChoiceTurnSkipOrder)
	defer func() { choiceGateDecisions[PendingChoiceTurnSkipOrder] = blocks }()
	if _, err := snap.Restore(); !errors.Is(err, ErrUnknownEffectKey) {
		t.Errorf("Restore without the kind = %v, want ErrUnknownEffectKey", err)
	}
}

// TestAStandingAnswerAnswersTheQuestion: ADR 0127's fixed engine key.
// Always is answer 0, Never answer 1, each with an EventAutoAnswer.
func TestAStandingAnswerAnswersTheQuestion(t *testing.T) {
	for _, tc := range []struct {
		rule AutoAnswer
		owed int
		call string
	}{
		{AutoAnswerAlways, 0, AutoAnswerCallYes},
		{AutoAnswerNever, 1, AutoAnswerCallNo},
	} {
		t.Run(string(tc.rule), func(t *testing.T) {
			g, _ := pausedTable(t, 1)
			// An empty library does not send "Always" to the hand: the
			// answer draws nothing (ADR 0127 owner decision 9 is about
			// draws).
			g.WithWriteLock(func() { g.Seats[0].Library.Cards = nil })
			if err := g.SetAutoAnswers(g.Seats[0].ID, map[string]AutoAnswer{TurnSkipOrderAutoAnswerKey: tc.rule}); err != nil {
				t.Fatalf("SetAutoAnswers: %v", err)
			}
			passTurn(t, g)
			id, chooser, ok := g.NextAutoAnswer()
			if !ok || chooser != g.Seats[0].ID || id != theQuestion(t, g).ID {
				t.Fatalf("NextAutoAnswer = %v %v %v", id, chooser, ok)
			}
			if _, err := g.AutoAnswer(id); err != nil {
				t.Fatalf("AutoAnswer: %v", err)
			}
			var ev *Event
			for i := range g.Events {
				if g.Events[i].Kind == EventAutoAnswer {
					ev = &g.Events[i]
				}
			}
			if ev == nil || ev.Call != tc.call || ev.AnswerKey != TurnSkipOrderAutoAnswerKey {
				t.Errorf("EventAutoAnswer = %+v, want call %q", ev, tc.call)
			}
			if g.TurnChangePending || g.Turn.ActiveSeat != 1 || g.TurnSkipsFor(g.Seats[0].ID) != tc.owed {
				t.Errorf("pending %v seat %d owing %d, want seat 1 owing %d",
					g.TurnChangePending, g.Turn.ActiveSeat, g.TurnSkipsFor(g.Seats[0].ID), tc.owed)
			}
		})
	}
}

// TestTheQuestionIsNeverAskedInsideAResolution: Final Fortune's end-step
// loss resolves in seat 0's extra turn; the next turn is seat 1's extra
// turn, seat 1 owes a skip, and a Trouble in Pairs skips seat 1's extra
// turns. The loss defers the rotation to the SBA pass after the
// resolution (ADR 0057 Decision 3), and only that rotation pauses.
func TestTheQuestionIsNeverAskedInsideAResolution(t *testing.T) {
	g := newFourPlayerActiveGame(t)
	g.WithWriteLock(func() { g.RegisterReplacementForTest(skipExtraTurnsOf(1)) })
	skipTurns(t, g, 1, uuid.Nil, 1)
	takeExtraTurns(t, g, 1, 1)
	ff := takeExtraTurns(t, g, 0, 1)
	g.WithWriteLock(func() {
		g.ScheduleDelayedTriggerForEffect(DelayedTrigger{
			Controller: g.Seats[0].ID, Label: "probe — you lose the game", At: StepEnd,
			OnExtraTurn: ff[0],
			Body: testBody(func(g *Game, _ *StackItem) error {
				if _, err := g.LoseTheGameForEffect(g.Seats[0].ID, uuid.Nil); err != nil {
					return err
				}
				if g.TurnChangePending {
					t.Error("the turn change paused inside the resolution")
				}
				return nil
			}),
		})
	})
	passTurn(t, g)
	if got := shapeOf(g); got != (turnShape{seat: 0, extra: true, round: 1}) {
		t.Fatalf("setup: %+v, want seat 0's Final Fortune turn", got)
	}
	for i := 0; i < 40 && !g.TurnChangePending; i++ {
		if _, err := g.AdvanceStep(); err != nil {
			t.Fatalf("AdvanceStep %d: %v", i, err)
		}
	}
	if !g.Seats[0].Eliminated {
		t.Fatal("seat 0 did not lose to its end-step trigger")
	}
	c := theQuestion(t, g)
	if c.Chooser != g.Seats[1].ID || c.midResolution || g.ActiveSeatLeftPending {
		t.Errorf("chooser %v midResolution %v leftPending %v", c.Chooser, c.midResolution, g.ActiveSeatLeftPending)
	}
	if len(g.StackMeta) != 0 {
		t.Errorf("%d items on the stack while asked", len(g.StackMeta))
	}
	answer(t, g, true)
	if g.TurnChangePending || g.Seats[g.Turn.ActiveSeat].Eliminated {
		t.Errorf("after the answer: pending %v, seat %d", g.TurnChangePending, g.Turn.ActiveSeat)
	}
}
