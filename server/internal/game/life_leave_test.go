package game

import (
	"errors"
	"testing"

	"github.com/google/uuid"
)

// life_leave_test.go is the #808 regression suite: the post-merge
// review of #806 found that the life-change continuation did not reach
// every terminal outcome once a player LEFT THE GAME mid-drain, plus a
// handful of smaller gaps in the replacement pipeline around it.
//
// Every departure here goes through Concede, not a UUID that was never
// seated: an eliminated seat stays in g.Seats, which is exactly what
// the original nil checks missed.

// recordTotals is the drain continuation every test here hands in.
func recordTotals(into *[]int) func(*Game, int) error {
	return func(_ *Game, n int) error {
		*into = append(*into, n)
		return nil
	}
}

// --- (1) the paused leg's player concedes ----------------------------

// TestConcedeDuringLifeOrderingPromptFinishesTheDrain — the regression.
// A's loss pauses on a CR 616 prompt addressed to A; A concedes instead
// of answering. The prompt is dropped with them, but the drain's later
// legs and its total still run: B loses 3 and the drain reports 3,
// which is what it reported before #806 made the legs sequential.
func TestConcedeDuringLifeOrderingPromptFinishesTheDrain(t *testing.T) {
	g := newActiveGameWithSeats(t, 4)
	a, b := g.Seats[1].ID, g.Seats[2].ID
	var total []int
	g.WithWriteLock(func() {
		g.RegisterReplacementForTest(plusOneLifeReplacement(a))
		g.RegisterReplacementForTest(doubleLifeReplacement(a))
		if err := g.LoseLifeEachThenForEffect(uuid.Nil, []uuid.UUID{a, b}, 3, recordTotals(&total)); err != nil {
			t.Fatalf("LoseLifeEachThenForEffect: %v", err)
		}
	})
	if len(g.PendingChoices) != 1 || g.PendingChoices[0].Chooser != a {
		t.Fatalf("want exactly A's CR 616 prompt open, got %d choices", len(g.PendingChoices))
	}

	if err := g.Concede(a); err != nil {
		t.Fatalf("Concede: %v", err)
	}

	if len(total) != 1 || total[0] != 3 {
		t.Fatalf("drain total = %v, want [3] — B's leg, and nothing from the player who left", total)
	}
	if got := g.Seats[2].Life; got != StartingLife-3 {
		t.Errorf("B life = %d, want %d — the later leg must still run", got, StartingLife-3)
	}
	if got := g.Seats[1].Life; got != StartingLife {
		t.Errorf("A life = %d, want %d — the conceded leg lands nothing", got, StartingLife)
	}
	if len(g.PendingChoices) != 0 {
		t.Errorf("%d choices still queued, want 0", len(g.PendingChoices))
	}
	if len(g.replacementsAppliedThisEvent) != 0 {
		t.Errorf("once-per-event map still holds %d entries after the dropped event", len(g.replacementsAppliedThisEvent))
	}
}

// TestConcedeDuringLastLegPromptStillReportsTheTotal — the same drop on
// the LAST leg: nothing follows it but the total, and the total is
// what the earlier legs lost.
func TestConcedeDuringLastLegPromptStillReportsTheTotal(t *testing.T) {
	g := newActiveGameWithSeats(t, 4)
	a, b := g.Seats[1].ID, g.Seats[2].ID
	var total []int
	g.WithWriteLock(func() {
		g.RegisterReplacementForTest(plusOneLifeReplacement(b))
		g.RegisterReplacementForTest(doubleLifeReplacement(b))
		if err := g.LoseLifeEachThenForEffect(uuid.Nil, []uuid.UUID{a, b}, 2, recordTotals(&total)); err != nil {
			t.Fatalf("LoseLifeEachThenForEffect: %v", err)
		}
	})
	if len(total) != 0 {
		t.Fatalf("total reported %v before B's paused leg settled", total)
	}
	if err := g.Concede(b); err != nil {
		t.Fatalf("Concede: %v", err)
	}
	if len(total) != 1 || total[0] != 2 {
		t.Fatalf("drain total = %v, want [2] — A's loss only", total)
	}
}

// --- (2) eliminated players are skipped (CR 800.4a) ------------------

// TestLoseLifeEachSkipsAPlayerWhoConcededDuringAnEarlierPrompt — the
// issue's second repro. A's prompt is open, B concedes, A answers.
// B's leg runs AFTER the answer, so it has to see B as gone: B loses
// nothing, and the caster is not paid for life a departed player never
// lost. Before the fix the total was 11 rather than 8.
func TestLoseLifeEachSkipsAPlayerWhoConcededDuringAnEarlierPrompt(t *testing.T) {
	g := newActiveGameWithSeats(t, 4)
	a, b := g.Seats[1].ID, g.Seats[2].ID
	var total []int
	g.WithWriteLock(func() {
		g.RegisterReplacementForTest(plusOneLifeReplacement(a))
		g.RegisterReplacementForTest(doubleLifeReplacement(a))
		if err := g.LoseLifeEachThenForEffect(uuid.Nil, []uuid.UUID{a, b}, 3, recordTotals(&total)); err != nil {
			t.Fatalf("LoseLifeEachThenForEffect: %v", err)
		}
	})
	if err := g.Concede(b); err != nil {
		t.Fatalf("Concede: %v", err)
	}
	answerOnlyPrompt(t, g)

	if len(total) != 1 || total[0] != 8 {
		t.Fatalf("drain total = %v, want [8] — (3+1)*2 from A, nothing from B", total)
	}
	if got := g.Seats[2].Life; got != StartingLife {
		t.Errorf("eliminated B life = %d, want %d", got, StartingLife)
	}
}

// TestLoseLifeEachSkipsAPlayerEliminatedBeforeTheDrain — the unpaused
// floor: a seat that conceded before the effect resolved is not a leg.
func TestLoseLifeEachSkipsAPlayerEliminatedBeforeTheDrain(t *testing.T) {
	g := newActiveGameWithSeats(t, 4)
	a, b, c := g.Seats[1].ID, g.Seats[2].ID, g.Seats[3].ID
	if err := g.Concede(b); err != nil {
		t.Fatalf("Concede: %v", err)
	}
	var total []int
	g.WithWriteLock(func() {
		if err := g.LoseLifeEachThenForEffect(uuid.Nil, []uuid.UUID{a, b, c}, 2, recordTotals(&total)); err != nil {
			t.Fatalf("LoseLifeEachThenForEffect: %v", err)
		}
	})
	if len(total) != 1 || total[0] != 4 {
		t.Fatalf("drain total = %v, want [4]", total)
	}
	if got := g.Seats[2].Life; got != StartingLife {
		t.Errorf("eliminated B life = %d, want %d", got, StartingLife)
	}
}

// TestLifeChangeForAnEliminatedPlayerIsANoOpWithZero — the single form.
// A gain aimed at a player who has left is not an error (the effect did
// nothing wrong) and moves nothing; the continuation hears zero.
func TestLifeChangeForAnEliminatedPlayerIsANoOpWithZero(t *testing.T) {
	g := newActiveGameWithSeats(t, 3)
	gone := g.Seats[1].ID
	if err := g.Concede(gone); err != nil {
		t.Fatalf("Concede: %v", err)
	}
	var got []int
	g.WithWriteLock(func() {
		if err := g.ChangePlayerLifeThenForEffect(uuid.Nil, gone, 5, recordTotals(&got)); err != nil {
			t.Fatalf("ChangePlayerLifeThenForEffect: %v", err)
		}
	})
	if len(got) != 1 || got[0] != 0 {
		t.Fatalf("continuation saw %v, want [0]", got)
	}
	if g.Seats[1].Life != StartingLife {
		t.Errorf("eliminated life = %d, want %d", g.Seats[1].Life, StartingLife)
	}
}

// TestAnsweredPromptForAnEliminatedPlayersLifeLandsNothing — the resume
// half of (2). The prompt is a CR 614.10 "may" controlled by C on A's
// life change, so it is C's to answer and A conceding does not drop it.
// When C answers, the event resumes for a player who has left:
// applyResolvedLifeChangeLocked must treat them as gone and run the
// continuation with zero.
func TestAnsweredPromptForAnEliminatedPlayersLifeLandsNothing(t *testing.T) {
	g := newActiveGameWithSeats(t, 4)
	a, c := g.Seats[1].ID, g.Seats[3].ID
	var got []int
	g.WithWriteLock(func() {
		g.RegisterReplacementForTest(optionalLifeReplacementControlledBy(c))
		if err := g.ChangePlayerLifeThenForEffect(uuid.Nil, a, -3, recordTotals(&got)); err != nil {
			t.Fatalf("ChangePlayerLifeThenForEffect: %v", err)
		}
	})
	if len(g.PendingChoices) != 1 || g.PendingChoices[0].Chooser != c {
		t.Fatalf("want C's optional prompt open, got %d choices", len(g.PendingChoices))
	}
	if err := g.Concede(a); err != nil {
		t.Fatalf("Concede: %v", err)
	}
	if len(got) != 0 {
		t.Fatalf("continuation ran %v when the affected player left — the prompt is C's, not A's", got)
	}
	p := g.PendingChoices[0]
	if err := g.ResolveOptionalReplacement(p.ID, c, true); err != nil {
		t.Fatalf("ResolveOptionalReplacement: %v", err)
	}
	if len(got) != 1 || got[0] != 0 {
		t.Fatalf("continuation saw %v, want [0]", got)
	}
	if g.Seats[1].Life != StartingLife {
		t.Errorf("eliminated A life = %d, want %d", g.Seats[1].Life, StartingLife)
	}
}

// TestConcedingChooserOfAMayOnSomeoneElsesLifeChangeLetsItLand — the
// other side of the drop: the player who left only owned the "may". A's
// life change still happens, with the "may" declined for the player who
// can no longer decide, and the continuation hears what moved.
func TestConcedingChooserOfAMayOnSomeoneElsesLifeChangeLetsItLand(t *testing.T) {
	g := newActiveGameWithSeats(t, 4)
	a, c := g.Seats[1].ID, g.Seats[3].ID
	var got []int
	g.WithWriteLock(func() {
		g.RegisterReplacementForTest(optionalLifeReplacementControlledBy(c))
		if err := g.ChangePlayerLifeThenForEffect(uuid.Nil, a, -3, recordTotals(&got)); err != nil {
			t.Fatalf("ChangePlayerLifeThenForEffect: %v", err)
		}
	})
	if err := g.Concede(c); err != nil {
		t.Fatalf("Concede: %v", err)
	}
	if len(got) != 1 || got[0] != -3 {
		t.Fatalf("continuation saw %v, want [-3] — the may declined, the loss landed", got)
	}
	if g.Seats[1].Life != StartingLife-3 {
		t.Errorf("A life = %d, want %d", g.Seats[1].Life, StartingLife-3)
	}
	if len(g.PendingChoices) != 0 {
		t.Errorf("%d choices still queued", len(g.PendingChoices))
	}
}

func optionalLifeReplacementControlledBy(controller uuid.UUID) ReplacementEffect {
	return ReplacementEffect{
		Watches:   []EventKind{EventChangeLife},
		AppliesTo: func(ev *ReplacementEvent, _ *Game, _ *Card) bool { return ev.Kind == RepEventLife },
		Optional:  true,
		Controller: func(*ReplacementEvent, *Game, *Card) uuid.UUID {
			return controller
		},
		Replace: func(ev *ReplacementEvent, _ *Game, _ *Card) error {
			ev.LifeDelta *= 10
			return nil
		},
		Label: "you may make it ten times that much",
	}
}

// --- the damage twin of (1) and (2) ----------------------------------

// TestConcedeDuringDamageOrderingPromptFinishesTheBatch — #807 built
// the damage batch on the same sequenced continuations, so it had the
// same drop. A's damage pauses, A concedes, B still takes 2.
func TestConcedeDuringDamageOrderingPromptFinishesTheBatch(t *testing.T) {
	g := newActiveGameWithSeats(t, 4)
	a, b := g.Seats[1].ID, g.Seats[2].ID
	var total []int
	g.WithWriteLock(func() {
		g.RegisterReplacementForTest(preventOneDamageTo(a))
		g.RegisterReplacementForTest(doubleDamageTo(a))
		if err := g.DealDamageEachThenForEffect(uuid.Nil, []uuid.UUID{a, b}, 2, recordTotals(&total)); err != nil {
			t.Fatalf("DealDamageEachThenForEffect: %v", err)
		}
	})
	if len(g.PendingChoices) != 1 {
		t.Fatalf("want A's CR 616 prompt open, got %d choices", len(g.PendingChoices))
	}
	if err := g.Concede(a); err != nil {
		t.Fatalf("Concede: %v", err)
	}
	if len(total) != 1 || total[0] != 2 {
		t.Fatalf("batch total = %v, want [2]", total)
	}
	if g.Seats[2].Life != StartingLife-2 {
		t.Errorf("B life = %d, want %d", g.Seats[2].Life, StartingLife-2)
	}
}

// TestDealDamageEachSkipsAPlayerWhoConcededDuringAnEarlierPrompt — and
// the damage twin of the eliminated-leg skip.
func TestDealDamageEachSkipsAPlayerWhoConcededDuringAnEarlierPrompt(t *testing.T) {
	g := newActiveGameWithSeats(t, 4)
	a, b := g.Seats[1].ID, g.Seats[2].ID
	var total []int
	g.WithWriteLock(func() {
		g.RegisterReplacementForTest(preventOneDamageTo(a))
		g.RegisterReplacementForTest(doubleDamageTo(a))
		if err := g.DealDamageEachThenForEffect(uuid.Nil, []uuid.UUID{a, b}, 3, recordTotals(&total)); err != nil {
			t.Fatalf("DealDamageEachThenForEffect: %v", err)
		}
	})
	if err := g.Concede(b); err != nil {
		t.Fatalf("Concede: %v", err)
	}
	answerOnlyPrompt(t, g)
	// Gather order: prevent 1 (2), then double (4).
	if len(total) != 1 || total[0] != 4 {
		t.Fatalf("batch total = %v, want [4] — A only", total)
	}
	if g.Seats[2].Life != StartingLife {
		t.Errorf("eliminated B life = %d, want %d", g.Seats[2].Life, StartingLife)
	}
}

// --- undo into an open prompt replays exactly ------------------------

// TestUndoIntoAnOpenPromptDoesNotReapplyAnEarlierEffect — an effect
// that applied ON ITS OWN before the prompt was queued is recorded only
// in the once-per-event map. Answering the prompt deletes that entry
// when the event settles, so a snapshot that did not carry its own copy
// replayed the answer with the mark gone and fired the effect again.
func TestUndoIntoAnOpenPromptDoesNotReapplyAnEarlierEffect(t *testing.T) {
	g := newActiveGameWithSeats(t, 3)
	a := g.Seats[1].ID
	var got []int
	g.WithWriteLock(func() {
		// Any loss: one more. Applies alone at -3, taking it to -4.
		g.RegisterReplacementForTest(plusOneLifeReplacement(a))
		// Two DIFFERENT effects that only switch on at -4 or worse,
		// so they prompt once the first has applied.
		g.RegisterReplacementForTest(gatedLifeReplacement(-4, "double at 4+", func(d int) int { return d * 2 }))
		g.RegisterReplacementForTest(gatedLifeReplacement(-4, "one more at 4+", func(d int) int { return d - 1 }))
		if err := g.ChangePlayerLifeThenForEffect(uuid.Nil, a, -3, recordTotals(&got)); err != nil {
			t.Fatalf("ChangePlayerLifeThenForEffect: %v", err)
		}
	})
	if len(g.PendingChoices) != 1 {
		t.Fatalf("want one CR 616 prompt, got %d", len(g.PendingChoices))
	}
	promptOpen := g.Clone()

	answerOnlyPrompt(t, g)
	first := g.Seats[1].Life
	// -3 → -4 (alone) → -8 → -9.
	if first != StartingLife-9 || len(got) != 1 || got[0] != -9 {
		t.Fatalf("first answer: life %d, continuation %v; want %d and [-9]", first, got, StartingLife-9)
	}

	got = nil
	g.WithWriteLock(func() { g.RestoreFrom(promptOpen) })
	answerOnlyPrompt(t, g)
	if g.Seats[1].Life != first {
		t.Errorf("replayed answer: life %d, want %d — the effect that applied before the prompt fired twice", g.Seats[1].Life, first)
	}
	if len(got) != 1 || got[0] != -9 {
		t.Errorf("replayed continuation %v, want [-9]", got)
	}
}

// gatedLifeReplacement applies to a life loss of at least -threshold
// (threshold is negative) and rewrites the delta with f.
func gatedLifeReplacement(threshold int, label string, f func(int) int) ReplacementEffect {
	return ReplacementEffect{
		Watches: []EventKind{EventChangeLife},
		AppliesTo: func(ev *ReplacementEvent, _ *Game, _ *Card) bool {
			return ev.Kind == RepEventLife && ev.LifeDelta <= threshold
		},
		Replace: func(ev *ReplacementEvent, _ *Game, _ *Card) error {
			ev.LifeDelta = f(ev.LifeDelta)
			return nil
		},
		Label: label,
	}
}

// halveLossReplacement halves a life loss, rounded toward zero loss
// ("lose half that much, rounded up").
func halveLossReplacement() ReplacementEffect {
	return ReplacementEffect{
		Watches: []EventKind{EventChangeLife},
		AppliesTo: func(ev *ReplacementEvent, _ *Game, _ *Card) bool {
			return ev.Kind == RepEventLife && ev.LifeDelta < 0
		},
		Replace: func(ev *ReplacementEvent, _ *Game, _ *Card) error {
			ev.LifeDelta = -((-ev.LifeDelta + 1) / 2)
			return nil
		},
		Label: "half that much",
	}
}

// --- CR 616.1f: re-check after each applied effect -------------------

// TestOrderingAnswerRechecksApplicabilityAfterEachEffect — halving a
// loss of 3 to 2 switches off a "lose 3 or more: one more" effect. The
// chosen order [halve, one more] must not fire the second on the
// strength of the gather the prompt was built from.
func TestOrderingAnswerRechecksApplicabilityAfterEachEffect(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0].ID
	g.WithWriteLock(func() {
		g.RegisterReplacementForTest(halveLossReplacement())
		g.RegisterReplacementForTest(gatedLifeReplacement(-3, "one more at 3+", func(d int) int { return d - 1 }))
		if err := g.ChangePlayerLifeForEffect(uuid.Nil, me, -3); err != nil {
			t.Fatalf("ChangePlayerLifeForEffect: %v", err)
		}
	})
	answerOnlyPrompt(t, g)
	if g.Seats[0].Life != StartingLife-2 {
		t.Errorf("life = %d, want %d — halved to 2, and the 3+ effect no longer applies (CR 616.1f)",
			g.Seats[0].Life, StartingLife-2)
	}
}

// TestCostPathRechecksApplicabilityAfterEachEffect — the same board on
// the cost path, which applies the gathered order without a prompt and
// used to apply all of it back to back.
func TestCostPathRechecksApplicabilityAfterEachEffect(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0].ID
	g.WithWriteLock(func() {
		g.RegisterReplacementForTest(halveLossReplacement())
		g.RegisterReplacementForTest(gatedLifeReplacement(-3, "one more at 3+", func(d int) int { return d - 1 }))
		if err := g.PayLifeForEffect(uuid.Nil, me, 3); err != nil {
			t.Fatalf("PayLifeForEffect: %v", err)
		}
	})
	if g.Seats[0].Life != StartingLife-2 {
		t.Errorf("life = %d, want %d — CR 616.1f on the cost path", g.Seats[0].Life, StartingLife-2)
	}
}

// TestEliminatedChooserPathRechecksApplicabilityAfterEachEffect — and
// on the gone-chooser escape: the prompt would be addressed to a player
// who has left, so the gathered order stands, one effect per pass.
func TestEliminatedChooserPathRechecksApplicabilityAfterEachEffect(t *testing.T) {
	g := newActiveGameWithSeats(t, 3)
	gone := g.Seats[1].ID
	if err := g.Concede(gone); err != nil {
		t.Fatalf("Concede: %v", err)
	}
	g.WithWriteLock(func() {
		g.RegisterReplacementForTest(halveLossReplacement())
		g.RegisterReplacementForTest(gatedLifeReplacement(-3, "one more at 3+", func(d int) int { return d - 1 }))
		// The sandbox verb reaches the pipeline for an eliminated
		// player (the effect entry point short-circuits), which is
		// the path that reads chooserGoneLocked.
		ev := &ReplacementEvent{Kind: RepEventLife, LifePlayer: gone, LifeDelta: -3}
		out, err := g.applyReplacementsLocked(ev)
		if err != nil {
			t.Fatalf("applyReplacementsLocked: %v", err)
		}
		if out == nil || out.LifeDelta != -2 {
			t.Fatalf("settled delta = %+v, want -2 (CR 616.1f)", out)
		}
		g.clearReplacementEventLocked(ev.ID)
	})
	if len(g.PendingChoices) != 0 {
		t.Errorf("%d prompts queued for a player who has left", len(g.PendingChoices))
	}
}

// --- a cancelled payment refuses the cost (CR 119.8, CR 614.17b) -----

// TestCancelledLifePaymentRefusesTheCost — "your life total can't
// change" makes a life payment impossible, not free.
func TestCancelledLifePaymentRefusesTheCost(t *testing.T) {
	g := newActiveGame(t)
	var err error
	g.WithWriteLock(func() {
		g.RegisterReplacementForTest(ReplacementEffect{
			Watches:    []EventKind{EventChangeLife},
			AppliesTo:  func(ev *ReplacementEvent, _ *Game, _ *Card) bool { return ev.Kind == RepEventLife },
			PureCancel: true,
			Replace: func(ev *ReplacementEvent, _ *Game, _ *Card) error {
				ev.Cancel()
				return nil
			},
			Label: "your life total can't change",
		})
		err = g.PayLifeForEffect(uuid.Nil, g.Seats[0].ID, 2)
	})
	if !errors.Is(err, ErrInvalidParam) {
		t.Fatalf("PayLifeForEffect = %v, want ErrInvalidParam — CR 119.8: the cost can't be paid", err)
	}
	if g.Seats[0].Life != StartingLife {
		t.Errorf("life = %d, want %d", g.Seats[0].Life, StartingLife)
	}
}

// --- the iteration cap on the resume paths ---------------------------

// lossChain registers n effects that each take a loss exactly one step
// further, starting at `from` (negative): the only way to run the
// apply-loop into its cap with a single applicable effect per pass.
func lossChain(g *Game, from, n int) {
	for i := 0; i < n; i++ {
		at := from - i
		g.RegisterReplacementForTest(ReplacementEffect{
			Watches: []EventKind{EventChangeLife},
			AppliesTo: func(ev *ReplacementEvent, _ *Game, _ *Card) bool {
				return ev.Kind == RepEventLife && ev.LifeDelta == at
			},
			Replace: func(ev *ReplacementEvent, _ *Game, _ *Card) error {
				ev.LifeDelta--
				return nil
			},
			Label: "chain",
		})
	}
}

// TestOrderingResumeAtTheIterationCapStillLandsAndRunsTheTail — the
// unpaused entry points let an event through as-is when the apply-loop
// hits its cap; the CR 616 resume used to return the error instead,
// dropping the event and its continuation on an already-dequeued
// prompt.
func TestOrderingResumeAtTheIterationCapStillLandsAndRunsTheTail(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0].ID
	var got []int
	g.WithWriteLock(func() {
		g.Seats[0].Life = 1000
		g.RegisterReplacementForTest(plusOneLifeReplacement(uuid.Nil))
		g.RegisterReplacementForTest(doubleLifeReplacement(uuid.Nil))
		// (3+1)*2 = 8 after the answer; the chain starts there.
		lossChain(g, -8, maxReplacementIters+8)
		if err := g.ChangePlayerLifeThenForEffect(uuid.Nil, me, -3, recordTotals(&got)); err != nil {
			t.Fatalf("ChangePlayerLifeThenForEffect: %v", err)
		}
	})
	answerOnlyPrompt(t, g)
	if len(got) != 1 {
		t.Fatalf("continuation ran %d times, want once", len(got))
	}
	if got[0] >= -8 || g.Seats[0].Life != 1000+got[0] {
		t.Errorf("life = %d, continuation %v — the capped event must land as it stood", g.Seats[0].Life, got)
	}
}

// TestOptionalResumeAtTheIterationCapStillLandsAndRunsTheTail — the
// same for the CR 614.10 "may" resume.
func TestOptionalResumeAtTheIterationCapStillLandsAndRunsTheTail(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0].ID
	var got []int
	g.WithWriteLock(func() {
		g.Seats[0].Life = 1000
		g.RegisterReplacementForTest(ReplacementEffect{
			Watches:   []EventKind{EventChangeLife},
			AppliesTo: func(ev *ReplacementEvent, _ *Game, _ *Card) bool { return ev.Kind == RepEventLife },
			Optional:  true,
			Replace: func(ev *ReplacementEvent, _ *Game, _ *Card) error {
				ev.LifeDelta -= 5
				return nil
			},
			Label: "you may make it five more",
		})
		lossChain(g, -8, maxReplacementIters+8)
		if err := g.ChangePlayerLifeThenForEffect(uuid.Nil, me, -3, recordTotals(&got)); err != nil {
			t.Fatalf("ChangePlayerLifeThenForEffect: %v", err)
		}
	})
	if len(g.PendingChoices) != 1 || g.PendingChoices[0].Kind != PendingChoiceOptionalReplacement {
		t.Fatalf("want one optional-replacement prompt, got %d choices", len(g.PendingChoices))
	}
	p := g.PendingChoices[0]
	if err := g.ResolveOptionalReplacement(p.ID, p.Chooser, true); err != nil {
		t.Fatalf("ResolveOptionalReplacement: %v", err)
	}
	if len(got) != 1 {
		t.Fatalf("continuation ran %d times, want once", len(got))
	}
	if got[0] >= -8 || g.Seats[0].Life != 1000+got[0] {
		t.Errorf("life = %d, continuation %v — the capped event must land as it stood", g.Seats[0].Life, got)
	}
}
