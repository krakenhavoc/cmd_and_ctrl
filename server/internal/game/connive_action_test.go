package game

import (
	"testing"

	"github.com/google/uuid"
)

// connive_action_test.go — #4589, the engine half: connive is a CR 614 event
// (KeywordActionConnive), opened once per conniving permanent at
// ConniveForEffect and performed by the caller's tail once the window has
// settled. The card half (Leader, Super-Genius, and every card that
// connives) is in cards/effects/leader_super_genius_test.go; the
// replacements below are test-injected, as keyword_action_test.go's are.

// conniveTailRan records what the test tail was told: the registered body
// captures nothing, so the test reads the last call from here.
var conniveTailRan struct {
	calls  int
	n      int
	amount int
	ids    []uuid.UUID
	carry  []uuid.UUID
}

var conniveTestTail = RegisterTail("test/connive-action", func(_ *Game, o TailOutcome) error {
	conniveTailRan.calls++
	conniveTailRan.n = o.Tail.N
	conniveTailRan.amount = o.Amount
	conniveTailRan.ids = o.IDs
	conniveTailRan.carry = o.Tail.Carry
	return nil
})

func resetConniveTail() { conniveTailRan.calls, conniveTailRan.n, conniveTailRan.amount = 0, 0, 0 }

// conniveExtraDrawReplacement is Leader's shape as a test injection.
func conniveExtraDrawReplacement(label string) ReplacementEffect {
	return ReplacementEffect{
		Watches: []EventKind{EventKeywordAction},
		AppliesTo: func(ev *ReplacementEvent, _ *Game, _ *Card) bool {
			return ev.Kind == RepEventKeywordAction && ev.KeywordAction == KeywordActionConnive
		},
		Replace: func(ev *ReplacementEvent, _ *Game, _ *Card) error {
			ev.ConniveExtraDraws++
			ev.KeywordActionCount = 1
			return nil
		},
		Label: label,
	}
}

func conniveTail(controller, creature uuid.UUID) KeyedTail {
	return KeyedTail{Ref: conniveTestTail, Player: controller, Carry: []uuid.UUID{creature}}
}

// With nothing watching the tail runs once, told the printed N and no extra
// draws, and the creature.
func TestConniveWindowRunsTheTailOnce(t *testing.T) {
	resetConniveTail()
	g := newActiveGame(t)
	me := g.Seats[0]
	creature := uuid.New()
	g.WithWriteLock(func() {
		if err := g.ConniveForEffect(me.ID, creature, 2, conniveTail(me.ID, creature)); err != nil {
			t.Fatalf("ConniveForEffect: %v", err)
		}
	})
	if conniveTailRan.calls != 1 || conniveTailRan.n != 2 || conniveTailRan.amount != 0 {
		t.Errorf("tail ran %d times with N %d and %d extra draws, want once, 2, 0", conniveTailRan.calls, conniveTailRan.n, conniveTailRan.amount)
	}
	if len(conniveTailRan.ids) != 1 || conniveTailRan.ids[0] != creature {
		t.Errorf("tail IDs = %v, want the conniving creature", conniveTailRan.ids)
	}
}

// A replacement's extra draws and rewritten count reach the tail.
func TestConniveReplacementReachesTheTail(t *testing.T) {
	resetConniveTail()
	g := newActiveGame(t)
	me := g.Seats[0]
	creature := uuid.New()
	g.WithWriteLock(func() {
		g.RegisterReplacementForTest(conniveExtraDrawReplacement("draw first"))
		if err := g.ConniveForEffect(me.ID, creature, 3, conniveTail(me.ID, creature)); err != nil {
			t.Fatalf("ConniveForEffect: %v", err)
		}
	})
	if conniveTailRan.calls != 1 || conniveTailRan.n != 1 || conniveTailRan.amount != 1 {
		t.Errorf("tail ran %d times with N %d and %d extra draws, want once, 1, 1", conniveTailRan.calls, conniveTailRan.n, conniveTailRan.amount)
	}
}

// A null replacement (CR 614.10) ends the connive; the tail still runs, told
// nothing connived (N is 0), so the rest of its sentence goes on.
func TestCancelledConniveStillRunsTheTailWithNothingDone(t *testing.T) {
	resetConniveTail()
	g := newActiveGame(t)
	me := g.Seats[0]
	creature := uuid.New()
	g.WithWriteLock(func() {
		g.RegisterReplacementForTest(keywordCancelReplacement(KeywordActionConnive))
		if err := g.ConniveForEffect(me.ID, creature, 1, conniveTail(me.ID, creature)); err != nil {
			t.Fatalf("ConniveForEffect: %v", err)
		}
	})
	if conniveTailRan.calls != 1 || conniveTailRan.n != 0 || conniveTailRan.amount != 0 {
		t.Errorf("tail ran %d times with N %d and %d extra draws, want once, 0, 0", conniveTailRan.calls, conniveTailRan.n, conniveTailRan.amount)
	}
}

// CR 701.50e: connive 0 is no connive event, and the tail is told nothing
// connived.
func TestConniveZeroOpensNoWindow(t *testing.T) {
	resetConniveTail()
	g := newActiveGame(t)
	me := g.Seats[0]
	creature := uuid.New()
	g.WithWriteLock(func() {
		g.RegisterReplacementForTest(conniveExtraDrawReplacement("draw first"))
		if err := g.ConniveForEffect(me.ID, creature, 0, conniveTail(me.ID, creature)); err != nil {
			t.Fatalf("ConniveForEffect: %v", err)
		}
	})
	if conniveTailRan.calls != 1 || conniveTailRan.n != 0 || conniveTailRan.amount != 0 {
		t.Errorf("tail ran %d times with N %d and %d extra draws, want once, 0, 0", conniveTailRan.calls, conniveTailRan.n, conniveTailRan.amount)
	}
}

// Two different replacements of one connive are a CR 616 ordering prompt:
// nothing runs until it is answered, the answered order decides what the tail
// is told, and it runs exactly once. "Draw first" sets the count to 1 and
// "one more" adds one, so the order is the difference between connive 1 and
// connive 2.
func TestConniveWindowThatPausesRunsTheTailOnTheAnswer(t *testing.T) {
	for _, tc := range []struct {
		name  string
		order []string
		n     int
	}{
		{"draw first then one more", []string{"draw first", "one more"}, 2},
		{"one more then draw first", []string{"one more", "draw first"}, 1},
	} {
		t.Run(tc.name, func(t *testing.T) {
			resetConniveTail()
			g := newActiveGame(t)
			me := g.Seats[0]
			creature := uuid.New()
			g.WithWriteLock(func() {
				g.RegisterReplacementForTest(conniveExtraDrawReplacement("draw first"))
				g.RegisterReplacementForTest(keywordCountReplacement(KeywordActionConnive, plusOneForTest, "one more"))
				if err := g.ConniveForEffect(me.ID, creature, 1, conniveTail(me.ID, creature)); err != nil {
					t.Fatalf("ConniveForEffect: %v", err)
				}
			})
			if conniveTailRan.calls != 0 {
				t.Fatalf("the tail ran %d times before the order was answered", conniveTailRan.calls)
			}
			if len(g.PendingChoices) != 1 || g.PendingChoices[0].Kind != PendingChoiceReplacementOrder {
				t.Fatalf("pending choices = %+v, want the one CR 616 prompt", g.PendingChoices)
			}
			p := g.PendingChoices[0]
			if err := g.ResolveReplacementOrder(p.ID, p.Chooser, orderByLabel(t, g, p, tc.order...)); err != nil {
				t.Fatalf("ResolveReplacementOrder: %v", err)
			}
			if conniveTailRan.calls != 1 || conniveTailRan.n != tc.n || conniveTailRan.amount != 1 {
				t.Errorf("tail ran %d times with N %d and %d extra draws, want once, %d, 1", conniveTailRan.calls, conniveTailRan.n, conniveTailRan.amount, tc.n)
			}
			if len(g.replacementsAppliedThisEvent) != 0 {
				t.Errorf("once-per-event map still holds %d entries", len(g.replacementsAppliedThisEvent))
			}
		})
	}
}
