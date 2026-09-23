package game

import (
	"testing"

	"github.com/google/uuid"
)

// counter_tail_test.go — #1282: a counter placement with a
// continuation. Every pausing test here builds a REAL pausing board:
// two different counter replacements that both apply to the placement,
// which is what makes the CR 614 window queue the CR 616 ordering
// prompt and return with nothing placed.

// counterProbe is a sourceless counter replacement for one counter
// kind, rewriting the delta of any PLACEMENT (positive delta) of it.
// Two of them with different rewrites is the pausing board: a
// Doubling Season beside a Hardened Scales, for any counter name.
func counterProbe(label, name string, rewrite func(int) int) ReplacementEffect {
	return ReplacementEffect{
		Watches: []EventKind{EventCounterPlaced},
		AppliesTo: func(ev *ReplacementEvent, _ *Game, _ *Card) bool {
			return ev.Kind == RepEventCounter && ev.CounterName == name && ev.CounterDelta > 0
		},
		Replace: func(ev *ReplacementEvent, _ *Game, _ *Card) error {
			ev.CounterDelta = rewrite(ev.CounterDelta)
			return nil
		},
		Label: label,
	}
}

func timesTwoCounters(n int) int { return n * 2 }
func plusOneCounter(n int) int   { return n + 1 }

// registerPausingCounterBoard puts the two replacements on the table
// for `name`: "twice" and "one more".
func registerPausingCounterBoard(g *Game, name string) {
	g.WithWriteLock(func() {
		g.RegisterReplacementForTest(counterProbe("twice", name, timesTwoCounters))
		g.RegisterReplacementForTest(counterProbe("one more", name, plusOneCounter))
	})
}

// answerCounterOrder answers the one open CR 616 prompt, "twice" first.
func answerCounterOrder(t *testing.T, g *Game) {
	t.Helper()
	if len(g.PendingChoices) != 1 {
		t.Fatalf("pending choices = %d, want the one CR 616 prompt", len(g.PendingChoices))
	}
	p := g.PendingChoices[0]
	if p.Kind != PendingChoiceReplacementOrder {
		t.Fatalf("prompt kind = %q, want %q", p.Kind, PendingChoiceReplacementOrder)
	}
	if err := g.ResolveReplacementOrder(p.ID, p.Chooser, orderByLabel(t, g, p, "twice", "one more")); err != nil {
		t.Fatalf("ResolveReplacementOrder: %v", err)
	}
}

// tailRecord is what a continuation saw when it ran.
type tailRecord struct {
	runs   int
	placed int
	seen   int // counters on the target at the moment the tail ran
}

func (r *tailRecord) then(target uuid.UUID, name string) func(g *Game, placed int) error {
	return func(g *Game, placed int) error {
		r.runs++
		r.placed = placed
		r.seen = counterOn(g, target, name)
		return nil
	}
}

// TestAddCounterThenWaitsForTheOrderPrompt is the issue: the rest of
// the effect runs when the counters LAND, not when the call returns.
func TestAddCounterThenWaitsForTheOrderPrompt(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0]
	bear := pushBear(g, me.ID)
	registerPausingCounterBoard(g, CounterPlusOne)

	var rec tailRecord
	g.WithWriteLock(func() {
		if err := g.AddCounterThenForEffect(bear, CounterPlusOne, 2, rec.then(bear, CounterPlusOne)); err != nil {
			t.Fatalf("AddCounterThenForEffect: %v", err)
		}
	})
	if rec.runs != 0 {
		t.Fatalf("the continuation ran %d time(s) with the CR 616 prompt still open — before the counters landed", rec.runs)
	}
	if got := counterOn(g, bear, CounterPlusOne); got > 0 {
		t.Fatalf("+1/+1 counters = %d before the order is answered, want none", got)
	}

	answerCounterOrder(t, g)

	if rec.runs != 1 {
		t.Fatalf("continuation ran %d times after the answer, want exactly once", rec.runs)
	}
	// ×2 then +1: 2 → 4 → 5.
	if rec.placed != 5 || rec.seen != 5 {
		t.Errorf("continuation saw placed=%d, on the card=%d; want 5 and 5", rec.placed, rec.seen)
	}
}

// TestAddCounterThenRunsInlineWithoutAPause: one replacement is no
// prompt, and the continuation runs before the call returns.
func TestAddCounterThenRunsInlineWithoutAPause(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0]
	bear := pushBear(g, me.ID)
	g.WithWriteLock(func() {
		g.RegisterReplacementForTest(counterProbe("twice", CounterPlusOne, timesTwoCounters))
	})

	var rec tailRecord
	g.WithWriteLock(func() {
		if err := g.AddCounterThenForEffect(bear, CounterPlusOne, 2, rec.then(bear, CounterPlusOne)); err != nil {
			t.Fatalf("AddCounterThenForEffect: %v", err)
		}
		if rec.runs != 1 || rec.placed != 4 || rec.seen != 4 {
			t.Errorf("inline continuation: runs=%d placed=%d seen=%d, want 1/4/4", rec.runs, rec.placed, rec.seen)
		}
	})
}

// TestACancelledCounterPlacementStillRunsTheTail: CR 614.10 with a
// null replacement places nothing, and the rest of the effect is told
// zero rather than left waiting.
func TestACancelledCounterPlacementStillRunsTheTail(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0]
	bear := pushBear(g, me.ID)
	g.WithWriteLock(func() {
		g.RegisterReplacementForTest(ReplacementEffect{
			Watches: []EventKind{EventCounterPlaced},
			AppliesTo: func(ev *ReplacementEvent, _ *Game, _ *Card) bool {
				return ev.Kind == RepEventCounter
			},
			Replace: func(ev *ReplacementEvent, _ *Game, _ *Card) error {
				ev.Cancel()
				return nil
			},
			Label: "no counters",
		})
	})
	var rec tailRecord
	g.WithWriteLock(func() {
		if err := g.AddCounterThenForEffect(bear, CounterPlusOne, 2, rec.then(bear, CounterPlusOne)); err != nil {
			t.Fatalf("AddCounterThenForEffect: %v", err)
		}
	})
	if rec.runs != 1 || rec.placed != 0 {
		t.Errorf("cancelled placement: runs=%d placed=%d, want 1/0", rec.runs, rec.placed)
	}
}

// TestAZeroCounterPlacementRunsTheTailWithZero: no event, but still a
// terminal outcome.
func TestAZeroCounterPlacementRunsTheTailWithZero(t *testing.T) {
	g := newActiveGame(t)
	bear := pushBear(g, g.Seats[0].ID)
	var rec tailRecord
	g.WithWriteLock(func() {
		if err := g.AddCounterThenForEffect(bear, CounterPlusOne, 0, rec.then(bear, CounterPlusOne)); err != nil {
			t.Fatalf("AddCounterThenForEffect: %v", err)
		}
	})
	if rec.runs != 1 || rec.placed != 0 {
		t.Errorf("zero placement: runs=%d placed=%d, want 1/0", rec.runs, rec.placed)
	}
}

// TestTheCounterTailIsToldZeroWhenTheTargetLeavesMidPrompt: the
// target is gone by the time the order is answered. The placement is
// dropped (ADR 0056 decision 3) and the continuation still runs, told
// zero, instead of waiting forever.
func TestTheCounterTailIsToldZeroWhenTheTargetLeavesMidPrompt(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0]
	bear := pushBear(g, me.ID)
	registerPausingCounterBoard(g, CounterPlusOne)

	var rec tailRecord
	g.WithWriteLock(func() {
		if err := g.AddCounterThenForEffect(bear, CounterPlusOne, 2, rec.then(bear, CounterPlusOne)); err != nil {
			t.Fatalf("AddCounterThenForEffect: %v", err)
		}
		// Out of the game entirely — not a graveyard, where a counter
		// could still land (CR 122.1 allows counters in any zone).
		g.Battlefield.Remove(bear)
	})
	answerCounterOrder(t, g)
	if rec.runs != 1 || rec.placed != 0 {
		t.Errorf("target gone: runs=%d placed=%d, want 1/0", rec.runs, rec.placed)
	}
}

// TestUndoAcrossAPausedCounterPlacementReplaysTheTail: a snapshot taken
// with the prompt open still carries the continuation, so undoing the
// answer and answering again runs the rest of the effect again rather
// than finding it consumed by the first run.
func TestUndoAcrossAPausedCounterPlacementReplaysTheTail(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0]
	bear := pushBear(g, me.ID)
	registerPausingCounterBoard(g, CounterPlusOne)

	var rec tailRecord
	g.WithWriteLock(func() {
		if err := g.AddCounterThenForEffect(bear, CounterPlusOne, 2, rec.then(bear, CounterPlusOne)); err != nil {
			t.Fatalf("AddCounterThenForEffect: %v", err)
		}
	})
	snapshot := g.Clone()
	answerCounterOrder(t, g)
	if rec.runs != 1 {
		t.Fatalf("first answer ran the continuation %d times, want 1", rec.runs)
	}

	g.RestoreFrom(snapshot)
	if got := counterOn(g, bear, CounterPlusOne); got > 0 {
		t.Fatalf("+1/+1 counters after the undo = %d, want none", got)
	}
	answerCounterOrder(t, g)
	if rec.runs != 2 || rec.seen != 5 {
		t.Errorf("replayed answer: runs=%d seen=%d, want 2 and 5", rec.runs, rec.seen)
	}
}

// TestAPausedPlacementCancelledByTheAnswerStillRunsTheTail: the
// ordering prompt is between a doubler and a replacement that cancels
// the placement outright. Answered, the window settles on "cancelled",
// and the resume's cancelled arm — not the landing — has to tell the
// continuation zero.
func TestAPausedPlacementCancelledByTheAnswerStillRunsTheTail(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0]
	bear := pushBear(g, me.ID)
	g.WithWriteLock(func() {
		g.RegisterReplacementForTest(counterProbe("twice", CounterPlusOne, timesTwoCounters))
		g.RegisterReplacementForTest(ReplacementEffect{
			Watches: []EventKind{EventCounterPlaced},
			AppliesTo: func(ev *ReplacementEvent, _ *Game, _ *Card) bool {
				return ev.Kind == RepEventCounter && ev.CounterDelta > 0
			},
			Replace: func(ev *ReplacementEvent, _ *Game, _ *Card) error {
				ev.Cancel()
				return nil
			},
			Label: "one more",
		})
	})
	var rec tailRecord
	g.WithWriteLock(func() {
		if err := g.AddCounterThenForEffect(bear, CounterPlusOne, 2, rec.then(bear, CounterPlusOne)); err != nil {
			t.Fatalf("AddCounterThenForEffect: %v", err)
		}
	})
	if rec.runs != 0 {
		t.Fatalf("continuation ran before the answer")
	}
	answerCounterOrder(t, g)
	if rec.runs != 1 || rec.placed != 0 {
		t.Errorf("cancelled on resume: runs=%d placed=%d, want 1/0", rec.runs, rec.placed)
	}
	if got := counterOn(g, bear, CounterPlusOne); got > 0 {
		t.Errorf("+1/+1 counters = %d, want none", got)
	}
}

// --- the migrated callers ----------------------------------------------

// TestEarthbendThenReadsTheLandAfterAPausedPlacement is Earthshape's
// shape: "Earthbend 3. Then … that land's power". On a board with two
// counter replacements the counters wait on the CR 616 prompt, and
// the rest of the sentence waits with them.
func TestEarthbendThenReadsTheLandAfterAPausedPlacement(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0]
	land := earthbendTestLand(t, g, me, "Forest")
	registerPausingCounterBoard(g, CounterPlusOne)

	runs, power := 0, -1
	g.WithWriteLock(func() {
		err := g.EarthbendThenForEffect(me.ID, uuid.Nil, land, 3, func(g *Game) error {
			runs++
			g.RecomputeLayersIfStaleLocked()
			if c, ok := g.battlefieldCardLocked(land); ok {
				power = c.CurrentPower()
			}
			return nil
		})
		if err != nil {
			t.Fatalf("EarthbendThenForEffect: %v", err)
		}
	})
	if runs != 0 {
		t.Fatalf("the rest of the sentence ran with the counters still owed (read power %d)", power)
	}
	answerCounterOrder(t, g)
	// ×2 then +1: 3 → 6 → 7, on a 0/0 body.
	if runs != 1 || power != 7 {
		t.Errorf("after the answer: runs=%d, that land's power=%d; want 1 and 7", runs, power)
	}
}

// TestEarthbendThenRunsWhenTheLandHasLeft: no land, no earthbend — and
// the sentence after "then" still runs.
func TestEarthbendThenRunsWhenTheLandHasLeft(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0]
	runs := 0
	g.WithWriteLock(func() {
		if err := g.EarthbendThenForEffect(me.ID, uuid.Nil, uuid.New(), 3, func(*Game) error {
			runs++
			return nil
		}); err != nil {
			t.Fatalf("EarthbendThenForEffect: %v", err)
		}
	})
	if runs != 1 {
		t.Errorf("continuation ran %d times, want 1", runs)
	}
}

// TestTheArmyYouAmassedIsReadAfterAPausedPlacement is Widespread
// Brutality's bug: CR 701.47c's clause read the Army at its PRE-amass
// power while the CR 616 prompt was open.
func TestTheArmyYouAmassedIsReadAfterAPausedPlacement(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0]
	army := seedArmy(g, me.ID, "Orc") // one +1/+1 counter, a 1/1
	registerPausingCounterBoard(g, CounterPlusOne)

	runs, power := 0, -1
	g.WithWriteLock(func() {
		err := g.AmassForEffect(me.ID, uuid.Nil, amassTestToken("Orc"), "Orc", 2, func(g *Game, chosen uuid.UUID) error {
			runs++
			g.RecomputeLayersIfStaleLocked()
			if c, ok := g.battlefieldCardLocked(chosen); ok {
				power = c.CurrentPower()
			}
			return nil
		})
		if err != nil {
			t.Fatalf("AmassForEffect: %v", err)
		}
	})
	if runs != 0 {
		t.Fatalf("\"the Army you amassed\" ran with the counters still owed (read power %d)", power)
	}
	answerCounterOrder(t, g)
	// 1 counter already, amass 2 → ×2 then +1 = 5 more.
	if runs != 1 || power != 6 {
		t.Errorf("after the answer: runs=%d, Army power=%d; want 1 and 6", runs, power)
	}
	if got := counterOn(g, army, CounterPlusOne); got != 6 {
		t.Errorf("Army counters = %d, want 6", got)
	}
}

// TestASagaWhoseEntryCounterPausesStillReachesChapterOne: the chapter
// check is the lore counter's continuation. Before #1282 it read the
// lore count on the next line — zero while the prompt was open — and
// the resume placed the counters without ever firing a chapter.
func TestASagaWhoseEntryCounterPausesStillReachesChapterOne(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0]
	saga := pushBattlefieldForTest(g, me.ID, "Test Saga", "Enchantment — Saga", "")
	registerPausingCounterBoard(g, CounterLore)

	g.WithWriteLock(func() { g.sagaEntersWithLoreCounterLocked(saga) })
	if n := countEventsOfKind(g, EventSagaChapter); n != 0 {
		t.Fatalf("%d chapter event(s) with the lore counter still owed", n)
	}
	answerCounterOrder(t, g)
	// ×2 then +1: 1 → 2 → 3 lore counters, so chapters I, II and III.
	if got := counterOn(g, saga, CounterLore); got != 3 {
		t.Fatalf("lore counters = %d, want 3", got)
	}
	if n := countEventsOfKind(g, EventSagaChapter); n != 3 {
		t.Errorf("chapter events after the answer = %d, want 3 (I, II and III)", n)
	}
}
