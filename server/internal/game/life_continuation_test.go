package game

import (
	"testing"

	"github.com/google/uuid"
)

// life_continuation_test.go is the #793 regression suite: what a life
// change owes its CALLER once the CR 614 window has settled it.
//
// #482 put every life change through that window, which means every
// life change can pause on a CR 616 ordering prompt. The callers that
// noticed were the ones that read the life total back on the next line
// — "each opponent loses X life, you gain life equal to the life lost
// this way" — because the read happened before the prompt was answered
// and reported that nothing had moved. life_tail.go carries the
// continuation instead, and this file pins the four things that has to
// be true of it: it runs with the REPLACED amount, it runs after a
// pause, it adds up across several players, and a cost never pauses at
// all.
//
// The scaffolding (runLifeScenario, the two-replacement board that
// makes CR 616 prompt) lives in life_tail_test.go and is shared.

// plusOneLifeReplacement and doubleLifeReplacement are the two
// DIFFERENT effects it takes to make a life event prompt: identical
// ones collapse without asking (#792) and pure cancels do too (#710).
// Gated on a player so a batch can have one opponent pause and another
// not.
func plusOneLifeReplacement(only uuid.UUID) ReplacementEffect {
	return ReplacementEffect{
		Watches: []EventKind{EventChangeLife},
		AppliesTo: func(ev *ReplacementEvent, _ *Game, _ *Card) bool {
			return ev.Kind == RepEventLife && (only == uuid.Nil || ev.LifePlayer == only)
		},
		Replace: func(ev *ReplacementEvent, _ *Game, _ *Card) error {
			if ev.LifeDelta < 0 {
				ev.LifeDelta--
			} else {
				ev.LifeDelta++
			}
			return nil
		},
		Label: "one more",
	}
}

func doubleLifeReplacement(only uuid.UUID) ReplacementEffect {
	return ReplacementEffect{
		Watches: []EventKind{EventChangeLife},
		AppliesTo: func(ev *ReplacementEvent, _ *Game, _ *Card) bool {
			return ev.Kind == RepEventLife && (only == uuid.Nil || ev.LifePlayer == only)
		},
		Replace: func(ev *ReplacementEvent, _ *Game, _ *Card) error {
			ev.LifeDelta *= 2
			return nil
		},
		Label: "twice that much",
	}
}

// answerOnlyPrompt answers the single queued CR 616 ordering prompt in
// the order it was offered, and fails if there isn't exactly one.
func answerOnlyPrompt(t *testing.T, g *Game) {
	t.Helper()
	if len(g.PendingChoices) != 1 {
		t.Fatalf("pending choices = %d, want exactly 1 CR 616 ordering prompt", len(g.PendingChoices))
	}
	p := g.PendingChoices[0]
	if p.Kind != PendingChoiceReplacementOrder {
		t.Fatalf("prompt kind = %q, want %q", p.Kind, PendingChoiceReplacementOrder)
	}
	if err := g.ResolveReplacementOrder(p.ID, p.Chooser, p.ReplacementEffectIDs); err != nil {
		t.Fatalf("ResolveReplacementOrder: %v", err)
	}
}

// --- (a) one change, paused, with a continuation ---------------------

// TestLifeContinuationRunsWithTheReplacedAmount — the unpaused floor.
// `then` is told what the window settled on, not what the caller asked
// for: a gain of 3 under a doubler is a gain of 6, and a card that
// mirrors it ("Sanguine Bond: that much life") has to mirror 6.
func TestLifeContinuationRunsWithTheReplacedAmount(t *testing.T) {
	g := newActiveGame(t)
	var got []int
	g.WithWriteLock(func() {
		g.RegisterReplacementForTest(doubleLifeReplacement(uuid.Nil))
		err := g.ChangePlayerLifeThenForEffect(uuid.Nil, g.Seats[0].ID, 3,
			func(_ *Game, applied int) error {
				got = append(got, applied)
				return nil
			})
		if err != nil {
			t.Fatalf("ChangePlayerLifeThenForEffect: %v", err)
		}
	})
	if len(got) != 1 || got[0] != 6 {
		t.Fatalf("continuation saw %v, want [6] — the REPLACED amount, once", got)
	}
	if g.Seats[0].Life != StartingLife+6 {
		t.Errorf("life = %d, want %d", g.Seats[0].Life, StartingLife+6)
	}
}

// TestLifeContinuationRunsAfterACR616Pause — the bug. Two different
// life replacements apply, the affected player is asked to order them,
// and the change does not land until they answer. The continuation has
// to wait with it and then be told the settled amount; before #793 the
// caller was told nothing at all and carried on with a life total that
// had not moved yet.
func TestLifeContinuationRunsAfterACR616Pause(t *testing.T) {
	g := newActiveGame(t)
	var got []int
	g.WithWriteLock(func() {
		g.RegisterReplacementForTest(plusOneLifeReplacement(uuid.Nil))
		g.RegisterReplacementForTest(doubleLifeReplacement(uuid.Nil))
		err := g.ChangePlayerLifeThenForEffect(uuid.Nil, g.Seats[0].ID, 3,
			func(_ *Game, applied int) error {
				got = append(got, applied)
				return nil
			})
		if err != nil {
			t.Fatalf("ChangePlayerLifeThenForEffect: %v", err)
		}
	})
	if len(got) != 0 {
		t.Fatalf("continuation ran %v before the prompt was answered — that is the read-back bug with extra steps", got)
	}
	if g.Seats[0].Life != StartingLife {
		t.Fatalf("life moved to %d before the prompt was answered", g.Seats[0].Life)
	}

	answerOnlyPrompt(t, g)

	// Gather order is registration order, so the answer as offered is
	// plus-one then double: (3+1)*2 = 8.
	if len(got) != 1 || got[0] != 8 {
		t.Fatalf("continuation saw %v, want [8] — (3+1)*2 after the ordering was answered", got)
	}
	if g.Seats[0].Life != StartingLife+8 {
		t.Errorf("life = %d, want %d", g.Seats[0].Life, StartingLife+8)
	}
}

// TestLifeContinuationRunsWithZeroWhenTheChangeIsReplacedAway —
// CR 614.10 with a null replacement. Nothing moved, so "you gain life
// equal to the life lost this way" gains nothing; the continuation is
// still told, because a drain adding up several opponents would
// otherwise wait forever on the one that lost nothing.
func TestLifeContinuationRunsWithZeroWhenTheChangeIsReplacedAway(t *testing.T) {
	g := newActiveGame(t)
	var got []int
	g.WithWriteLock(func() {
		g.RegisterReplacementForTest(ReplacementEffect{
			Watches:   []EventKind{EventChangeLife},
			AppliesTo: func(ev *ReplacementEvent, _ *Game, _ *Card) bool { return ev.Kind == RepEventLife },
			Replace: func(ev *ReplacementEvent, _ *Game, _ *Card) error {
				ev.Cancel()
				return nil
			},
			Label: "your life total can't change",
		})
		err := g.ChangePlayerLifeThenForEffect(uuid.Nil, g.Seats[0].ID, -4,
			func(_ *Game, applied int) error {
				got = append(got, applied)
				return nil
			})
		if err != nil {
			t.Fatalf("ChangePlayerLifeThenForEffect: %v", err)
		}
	})
	if len(got) != 1 || got[0] != 0 {
		t.Fatalf("continuation saw %v, want [0]", got)
	}
	if g.Seats[0].Life != StartingLife {
		t.Errorf("life = %d, want %d — the loss was replaced away", g.Seats[0].Life, StartingLife)
	}
}

// --- (b) the batch form ----------------------------------------------

// TestLoseLifeEachThenSumsTheAppliedAmounts — the unpaused floor for
// the batch. One opponent is under a life-loss doubler and the other
// is not, so "the life lost this way" is 2+4 = 6 rather than 2×2.
func TestLoseLifeEachThenSumsTheAppliedAmounts(t *testing.T) {
	g := newActiveGameWithSeats(t, 3)
	me, a, b := g.Seats[0].ID, g.Seats[1].ID, g.Seats[2].ID
	var total []int
	g.WithWriteLock(func() {
		g.RegisterReplacementForTest(doubleLifeReplacement(b))
		err := g.LoseLifeEachThenForEffect(uuid.Nil, []uuid.UUID{a, b}, 2,
			func(_ *Game, lost int) error {
				total = append(total, lost)
				return nil
			})
		if err != nil {
			t.Fatalf("LoseLifeEachThenForEffect: %v", err)
		}
	})
	if len(total) != 1 || total[0] != 6 {
		t.Fatalf("total lost = %v, want [6] — 2 from one opponent and 4 from the doubled one", total)
	}
	if g.Seats[1].Life != StartingLife-2 || g.Seats[2].Life != StartingLife-4 {
		t.Errorf("lives = %d/%d, want %d/%d", g.Seats[1].Life, g.Seats[2].Life, StartingLife-2, StartingLife-4)
	}
	if g.Seats[0].Life != StartingLife {
		t.Errorf("the drainer's own life moved to %d — the batch drains the listed players only", g.Seats[0].Life)
	}
	_ = me
}

// TestLoseLifeEachThenWaitsForAPausedLeg — (b) proper: one opponent's
// loss pauses on a CR 616 prompt and the other's does not. The total
// has to be the sum of what BOTH really lost, which is the number the
// pre-#793 code could not produce: it read the paused opponent's life
// total back before the prompt was answered and counted zero.
func TestLoseLifeEachThenWaitsForAPausedLeg(t *testing.T) {
	g := newActiveGameWithSeats(t, 3)
	a, b := g.Seats[1].ID, g.Seats[2].ID
	var total []int
	g.WithWriteLock(func() {
		// Two different replacements on `a` only: that one prompts.
		g.RegisterReplacementForTest(plusOneLifeReplacement(a))
		g.RegisterReplacementForTest(doubleLifeReplacement(a))
		err := g.LoseLifeEachThenForEffect(uuid.Nil, []uuid.UUID{a, b}, 3,
			func(_ *Game, lost int) error {
				total = append(total, lost)
				return nil
			})
		if err != nil {
			t.Fatalf("LoseLifeEachThenForEffect: %v", err)
		}
	})
	if len(total) != 0 {
		t.Fatalf("the batch reported %v before its paused leg landed", total)
	}

	answerOnlyPrompt(t, g)

	// a: -3 → -4 → -8. b: -3, untouched. Total lost = 11.
	if len(total) != 1 || total[0] != 11 {
		t.Fatalf("total lost = %v, want [11] — 8 from the paused opponent plus 3 from the other", total)
	}
	if g.Seats[1].Life != StartingLife-8 {
		t.Errorf("paused opponent life = %d, want %d", g.Seats[1].Life, StartingLife-8)
	}
	if g.Seats[2].Life != StartingLife-3 {
		t.Errorf("unpaused opponent life = %d, want %d", g.Seats[2].Life, StartingLife-3)
	}
}

// TestLoseLifeEachThenSkipsPlayersWhoAreGone — a batch over a seat that
// has left reports what the rest lost rather than stalling on it.
func TestLoseLifeEachThenSkipsPlayersWhoAreGone(t *testing.T) {
	g := newActiveGame(t)
	gone := uuid.New()
	var total []int
	g.WithWriteLock(func() {
		err := g.LoseLifeEachThenForEffect(uuid.Nil, []uuid.UUID{gone, g.Seats[1].ID, gone}, 2,
			func(_ *Game, lost int) error {
				total = append(total, lost)
				return nil
			})
		if err != nil {
			t.Fatalf("LoseLifeEachThenForEffect: %v", err)
		}
	})
	if len(total) != 1 || total[0] != 2 {
		t.Fatalf("total lost = %v, want [2]", total)
	}
}

// --- (c) paying life as a cost ---------------------------------------

// TestPayLifeRunsTheLifeLossWindow — the rule #793 guessed wrong and
// the CR settles. CR 119.4: "paying an amount of life is the same as
// losing that much life", so a life-LOSS replacement sees a cost being
// paid and a Thoughtseize under Bloodletter of Aclazotz really does
// cost four.
func TestPayLifeRunsTheLifeLossWindow(t *testing.T) {
	g := newActiveGame(t)
	g.WithWriteLock(func() {
		g.RegisterReplacementForTest(doubleLifeReplacement(uuid.Nil))
		if err := g.PayLifeForEffect(uuid.Nil, g.Seats[0].ID, 2); err != nil {
			t.Fatalf("PayLifeForEffect: %v", err)
		}
	})
	if g.Seats[0].Life != StartingLife-4 {
		t.Errorf("life = %d, want %d — CR 119.4 makes the payment a life loss", g.Seats[0].Life, StartingLife-4)
	}
}

// TestPayLifeIsNotLifeGain — the other half of the rule, and the half
// the issue was actually worried about. Rhox Faithmender and
// Alhammarret's Archive watch for a POSITIVE delta, so a payment never
// reaches them; nothing here needs a special case, but a future
// refactor that flipped the sign of a payment would be caught.
func TestPayLifeIsNotLifeGain(t *testing.T) {
	g := newActiveGame(t)
	fired := 0
	g.WithWriteLock(func() {
		g.RegisterReplacementForTest(ReplacementEffect{
			Watches: []EventKind{EventChangeLife},
			AppliesTo: func(ev *ReplacementEvent, _ *Game, _ *Card) bool {
				return ev.Kind == RepEventLife && ev.LifeDelta > 0
			},
			Replace: func(ev *ReplacementEvent, _ *Game, _ *Card) error {
				fired++
				ev.LifeDelta *= 2
				return nil
			},
			Label: "gain twice that much",
		})
		if err := g.PayLifeForEffect(uuid.Nil, g.Seats[0].ID, 3); err != nil {
			t.Fatalf("PayLifeForEffect: %v", err)
		}
	})
	if fired != 0 {
		t.Errorf("a lifegain replacement fired %d times on a payment, want 0", fired)
	}
	if g.Seats[0].Life != StartingLife-3 {
		t.Errorf("life = %d, want %d", g.Seats[0].Life, StartingLife-3)
	}
}

// TestPayLifeNeverPausesOnACR616Prompt — the constraint the cost path
// exists for. Two different life-loss replacements would ordinarily be
// a CR 616 ordering question; a cost cannot stop to ask one, because
// CR 601.2h pays a spell's costs as one indivisible step and a paused
// prompt leaves the spell on the stack half paid for. The gathered
// order stands, the payment is complete when the call returns, and the
// table is not waiting on anything.
func TestPayLifeNeverPausesOnACR616Prompt(t *testing.T) {
	g := newActiveGame(t)
	g.WithWriteLock(func() {
		g.RegisterReplacementForTest(plusOneLifeReplacement(uuid.Nil))
		g.RegisterReplacementForTest(doubleLifeReplacement(uuid.Nil))
		if err := g.PayLifeForEffect(uuid.Nil, g.Seats[0].ID, 2); err != nil {
			t.Fatalf("PayLifeForEffect: %v", err)
		}
	})
	if len(g.PendingChoices) != 0 {
		t.Fatalf("a cost payment queued %d prompts, want 0 — a half-paid cost cannot be rewound",
			len(g.PendingChoices))
	}
	// Gather order: one more (-3), then double (-6).
	if g.Seats[0].Life != StartingLife-6 {
		t.Errorf("life = %d, want %d — both replacements applied, in gather order",
			g.Seats[0].Life, StartingLife-6)
	}
}

// TestPayLifeSkipsAReplacementThatWouldAskAQuestion — the same rule for
// a CR 614.10 "may". The payment cannot pause to ask, and firing the
// effect blind would answer for its controller in the direction that
// favours it, so it is skipped un-applied: weaker than printed, never
// stronger.
func TestPayLifeSkipsAReplacementThatWouldAskAQuestion(t *testing.T) {
	g := newActiveGame(t)
	g.WithWriteLock(func() {
		g.RegisterReplacementForTest(ReplacementEffect{
			Watches:   []EventKind{EventChangeLife},
			AppliesTo: func(ev *ReplacementEvent, _ *Game, _ *Card) bool { return ev.Kind == RepEventLife },
			Optional:  true,
			Replace: func(ev *ReplacementEvent, _ *Game, _ *Card) error {
				ev.LifeDelta *= 5
				return nil
			},
			Label: "you may lose five times that much",
		})
		if err := g.PayLifeForEffect(uuid.Nil, g.Seats[0].ID, 2); err != nil {
			t.Fatalf("PayLifeForEffect: %v", err)
		}
	})
	if len(g.PendingChoices) != 0 {
		t.Fatalf("a cost payment queued %d prompts, want 0", len(g.PendingChoices))
	}
	if g.Seats[0].Life != StartingLife-2 {
		t.Errorf("life = %d, want %d — the optional replacement was skipped un-applied",
			g.Seats[0].Life, StartingLife-2)
	}
}

// TestPayLifeRefusesWhatCannotBePaid — CR 119.4: a player may pay life
// only from a total at least as large as the payment. Every caller
// checks before it starts paying; this is the backstop, and it must not
// leave a partial payment behind.
func TestPayLifeRefusesWhatCannotBePaid(t *testing.T) {
	g := newActiveGame(t)
	g.WithWriteLock(func() {
		g.Seats[0].Life = 3
		if err := g.PayLifeForEffect(uuid.Nil, g.Seats[0].ID, 4); err == nil {
			t.Fatal("PayLifeForEffect = nil, want an error — CR 119.4 forbids paying 4 from 3")
		}
	})
	if g.Seats[0].Life != 3 {
		t.Errorf("life = %d, want 3 — a refused payment pays nothing", g.Seats[0].Life)
	}
}

// --- (d) undo across the pause ---------------------------------------

// TestUndoAcrossALifePauseReplaysTheSameWay — the continuation is held
// on an unserialisable frame, so the undo stack has to be able to rewind
// a game that is sitting on one and have the replay come out the same.
//
// Two rewinds are checked, because they fail differently. Rewinding to
// BEFORE the change drops the prompt and the continuation together.
// Rewinding to WHILE THE PROMPT IS OPEN keeps them, and answering the
// prompt a second time has to produce the same total as the first —
// which is why the batch carries its running total forward by value
// instead of accumulating into a shared counter that the rewind could
// not undo.
func TestUndoAcrossALifePauseReplaysTheSameWay(t *testing.T) {
	g := newActiveGameWithSeats(t, 3)
	a, b := g.Seats[1].ID, g.Seats[2].ID
	var total []int
	drain := func() {
		err := g.LoseLifeEachThenForEffect(uuid.Nil, []uuid.UUID{a, b}, 3,
			func(_ *Game, lost int) error {
				total = append(total, lost)
				return nil
			})
		if err != nil {
			t.Fatalf("LoseLifeEachThenForEffect: %v", err)
		}
	}

	g.WithWriteLock(func() {
		g.RegisterReplacementForTest(plusOneLifeReplacement(a))
		g.RegisterReplacementForTest(doubleLifeReplacement(a))
	})

	// --- rewind to before the drain ---
	beforeDrain := g.Clone()
	g.WithWriteLock(drain)
	if len(g.PendingChoices) != 1 {
		t.Fatalf("pending choices = %d, want the CR 616 prompt", len(g.PendingChoices))
	}
	g.WithWriteLock(func() { g.RestoreFrom(beforeDrain) })
	if len(g.PendingChoices) != 0 {
		t.Fatalf("%d prompts survived the rewind to before the drain", len(g.PendingChoices))
	}
	if g.Seats[1].Life != StartingLife || g.Seats[2].Life != StartingLife {
		t.Fatalf("lives = %d/%d after rewinding to before the drain, want both %d",
			g.Seats[1].Life, g.Seats[2].Life, StartingLife)
	}

	// --- rewind to while the prompt is open, then answer twice ---
	total = nil
	g.WithWriteLock(drain)
	promptOpen := g.Clone()
	answerOnlyPrompt(t, g)
	first := observeLife(g, &lifeEventRecorder{})
	if len(total) != 1 || total[0] != 11 {
		t.Fatalf("first run total = %v, want [11]", total)
	}

	total = nil
	g.WithWriteLock(func() { g.RestoreFrom(promptOpen) })
	if g.Seats[1].Life != StartingLife || g.Seats[2].Life != StartingLife {
		t.Fatalf("lives = %d/%d after rewinding into the open prompt, want both %d",
			g.Seats[1].Life, g.Seats[2].Life, StartingLife)
	}
	answerOnlyPrompt(t, g)
	if len(total) != 1 || total[0] != 11 {
		t.Fatalf("replayed total = %v, want [11] — the same answer must give the same total", total)
	}
	if second := observeLife(g, &lifeEventRecorder{}); !equalInts(second.Life, first.Life) {
		t.Errorf("replayed lives %v differ from the first run's %v", second.Life, first.Life)
	}
}

func equalInts(a, b []int) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}
