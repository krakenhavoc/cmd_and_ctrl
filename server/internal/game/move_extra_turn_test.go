package game

import (
	"errors"
	"testing"

	"github.com/google/uuid"
)

// move_extra_turn_test.go — ADR 0188 Decision 2 (#3560): an extra turn
// that is part of a replaced move ("instead exile it and take an extra
// turn after this one") is declared on the move event and queued once
// the card has landed, never inside the Replace callback.

// exileAndExtraTurnInstead is Ugin's Nexus' second ability as a fixture:
// "if a creature would die, instead exile it and `taker` takes an extra
// turn after this one".
func exileAndExtraTurnInstead(taker uuid.UUID) ReplacementEffect {
	return ReplacementEffect{
		Watches: []EventKind{EventZoneMove},
		AppliesTo: func(ev *ReplacementEvent, g *Game, _ *Card) bool {
			if ev.Kind != RepEventMove || ev.OldZone != ZoneBattlefield || ev.NewZone != ZoneGraveyard {
				return false
			}
			c, ok := g.LookupCardForEffect(ev.CardID)
			return ok && c.IsCreature()
		},
		Replace: func(ev *ReplacementEvent, g *Game, _ *Card) error {
			ev.NewZone = ZoneExile
			ev.NewZoneOwner = uuid.Nil
			return g.TakeExtraTurnAfterReplacedMoveForEffect(ev, uuid.Nil, taker)
		},
		Label: "exile instead, and take an extra turn",
	}
}

func moveTurnBoard(t *testing.T) (*Game, *Player, uuid.UUID) {
	t.Helper()
	g := newActiveGameWithSeats(t, 3)
	me := g.Seats[0]
	bear := pushBear(g, g.Seats[1].ID)
	g.WithWriteLock(func() { g.RegisterReplacementForTest(exileAndExtraTurnInstead(me.ID)) })
	return g, me, bear
}

func queuedTurns(g *Game) int {
	n := 0
	g.WithWriteLock(func() { n = len(g.ExtraTurns) })
	return n
}

// The card is exiled first, then the turn is queued, then the move's
// continuation runs and already sees the queued turn.
func TestMoveExtraTurnLandsAfterTheMoveAndBeforeTheContinuation(t *testing.T) {
	g, _, bear := moveTurnBoard(t)
	var inExile, queued []bool
	g.WithWriteLock(func() {
		err := g.DestroyPermanentsKeyedForEffect([]uuid.UUID{bear}, hookTail(func(g *Game, o TailOutcome) error {
			inExile = append(inExile, g.Exile.Contains(bear))
			queued = append(queued, len(g.ExtraTurns) == 1)
			return nil
		}))
		if err != nil {
			t.Fatalf("destroy: %v", err)
		}
	})
	if len(inExile) != 1 || !inExile[0] || !queued[0] {
		t.Fatalf("continuation saw exiled=%v turnQueued=%v, want one run seeing both", inExile, queued)
	}
	exileAt, turnAt := -1, -1
	for i, ev := range g.Events {
		if ev.Kind == EventZoneMove && ev.CardID == bear && ev.NewZone == ZoneExile {
			exileAt = i
		}
		if ev.Kind == EventExtraTurnAdded {
			turnAt = i
		}
	}
	if exileAt < 0 || turnAt < exileAt {
		t.Errorf("exile at event %d, extra turn at %d: the turn must follow the exile", exileAt, turnAt)
	}
}

// A Replace callback must not emit an event: nothing is queued until the
// card has moved.
func TestMoveExtraTurnIsNotQueuedInsideReplace(t *testing.T) {
	g, _, bear := moveTurnBoard(t)
	g.WithWriteLock(func() {
		g.RegisterReplacementForTest(ReplacementEffect{
			Watches: []EventKind{EventZoneMove},
			AppliesTo: func(ev *ReplacementEvent, _ *Game, _ *Card) bool {
				return ev.Kind == RepEventMove && ev.CardID == bear && ev.NewZone == ZoneExile
			},
			Replace: func(ev *ReplacementEvent, g *Game, _ *Card) error {
				if len(g.ExtraTurns) != 0 {
					t.Errorf("the turn was queued inside the Replace callback")
				}
				return nil
			},
			Label: "observer",
		})
	})
	g.WithWriteLock(func() { _ = g.DestroyPermanentForEffect(bear) })
	if got := queuedTurns(g); got != 1 {
		t.Fatalf("%d turns queued, want 1", got)
	}
}

// An instruction declared before another replacement cancels the move
// still happens (CR 614.6).
func TestMoveExtraTurnSurvivesALaterCancel(t *testing.T) {
	g, _, bear := moveTurnBoard(t)
	g.WithWriteLock(func() {
		g.RegisterReplacementForTest(ReplacementEffect{
			Watches: []EventKind{EventZoneMove},
			AppliesTo: func(ev *ReplacementEvent, _ *Game, _ *Card) bool {
				return ev.Kind == RepEventMove && ev.CardID == bear && ev.NewZone == ZoneExile
			},
			Replace: func(ev *ReplacementEvent, _ *Game, _ *Card) error {
				ev.Canceled = true
				return nil
			},
			PureCancel: true,
			Label:      "it stays",
		})
	})
	ran := 0
	g.WithWriteLock(func() {
		_ = g.DestroyPermanentsKeyedForEffect([]uuid.UUID{bear}, hookTail(func(_ *Game, o TailOutcome) error {
			ran++
			return nil
		}))
	})
	if !g.Battlefield.Contains(bear) {
		t.Fatal("the cancelled move moved the creature")
	}
	if got := queuedTurns(g); got != 1 || ran != 1 {
		t.Fatalf("%d turns queued, continuation ran %d times; want 1 and 1", got, ran)
	}
}

// The same through the zone router (a bounce is not a destroy), whose
// cancel path is a different site: the declared turn is still queued.
func TestMoveExtraTurnSurvivesACancelOnTheRouterPath(t *testing.T) {
	g := newActiveGameWithSeats(t, 3)
	me := g.Seats[0]
	bear := pushBear(g, g.Seats[1].ID)
	g.WithWriteLock(func() {
		g.RegisterReplacementForTest(ReplacementEffect{
			Watches: []EventKind{EventZoneMove},
			AppliesTo: func(ev *ReplacementEvent, _ *Game, _ *Card) bool {
				return ev.Kind == RepEventMove && ev.CardID == bear && ev.NewZone == ZoneHand
			},
			Replace: func(ev *ReplacementEvent, g *Game, _ *Card) error {
				ev.NewZone = ZoneExile
				ev.NewZoneOwner = uuid.Nil
				return g.TakeExtraTurnAfterReplacedMoveForEffect(ev, uuid.Nil, me.ID)
			},
			Label: "exile instead of bouncing, and take an extra turn",
		})
		g.RegisterReplacementForTest(ReplacementEffect{
			Watches: []EventKind{EventZoneMove},
			AppliesTo: func(ev *ReplacementEvent, _ *Game, _ *Card) bool {
				return ev.Kind == RepEventMove && ev.CardID == bear && ev.NewZone == ZoneExile
			},
			Replace: func(ev *ReplacementEvent, _ *Game, _ *Card) error {
				ev.Canceled = true
				return nil
			},
			PureCancel: true,
			Label:      "it stays",
		})
	})
	g.WithWriteLock(func() {
		if err := g.BounceToHandForEffect(bear); err != nil {
			t.Fatalf("bounce: %v", err)
		}
	})
	if !g.Battlefield.Contains(bear) {
		t.Fatal("the cancelled move moved the creature")
	}
	if got := queuedTurns(g); got != 1 {
		t.Fatalf("%d turns queued, want 1", got)
	}
}

// A move that declares its extra turn and THEN pauses on a "may"
// replacement that only applies once the destination is exile queues
// nothing while the question is open; undoing into the question and
// answering again queues exactly one turn each time, because the
// snapshot has its own copy of the declared instruction.
func TestMoveExtraTurnAfterAPausedMoveAndUndo(t *testing.T) {
	g, _, bear := moveTurnBoard(t)
	g.WithWriteLock(func() {
		g.RegisterReplacementForTest(ReplacementEffect{
			Watches: []EventKind{EventZoneMove},
			AppliesTo: func(ev *ReplacementEvent, _ *Game, _ *Card) bool {
				return ev.Kind == RepEventMove && ev.CardID == bear && ev.NewZone == ZoneExile
			},
			Replace: func(ev *ReplacementEvent, _ *Game, _ *Card) error {
				ev.ExileCounter = "test"
				return nil
			},
			Optional: true,
			Label:    "a may replacement on the exile",
		})
	})
	g.WithWriteLock(func() { _ = g.DestroyPermanentForEffect(bear) })
	if len(g.PendingChoices) != 1 || g.PendingChoices[0].Kind != PendingChoiceOptionalReplacement {
		t.Fatalf("want the move's may-prompt, have %d choices", len(g.PendingChoices))
	}
	if got := queuedTurns(g); got != 0 {
		t.Fatalf("%d turns queued while the move was paused", got)
	}
	promptOpen := g.Clone()

	answer := func() {
		t.Helper()
		p := g.PendingChoices[0]
		if err := g.ResolveOptionalReplacement(p.ID, p.Chooser, true); err != nil {
			t.Fatalf("ResolveOptionalReplacement: %v", err)
		}
	}
	answer()
	if !g.Exile.Contains(bear) {
		t.Fatal("the answered move did not exile the creature")
	}
	if got := queuedTurns(g); got != 1 {
		t.Fatalf("%d turns queued after the answer, want 1", got)
	}

	g.WithWriteLock(func() { g.RestoreFrom(promptOpen) })
	if got := queuedTurns(g); got != 0 {
		t.Fatalf("%d turns queued after rewinding into the prompt", got)
	}
	answer()
	if got := queuedTurns(g); got != 1 {
		t.Fatalf("%d turns queued after the replayed answer, want 1", got)
	}
}

// A taker who has left the game gets no turn (CR 800.4a).
func TestMoveExtraTurnForAPlayerWhoLeftIsDropped(t *testing.T) {
	g, me, bear := moveTurnBoard(t)
	g.WithWriteLock(func() { me.Eliminated = true })
	g.WithWriteLock(func() { _ = g.DestroyPermanentForEffect(bear) })
	if got := queuedTurns(g); got != 0 {
		t.Fatalf("%d turns queued for an eliminated player", got)
	}
}

// Only a zone change may declare one.
func TestMoveExtraTurnRefusesANonMoveEvent(t *testing.T) {
	g := newActiveGameWithSeats(t, 3)
	var err error
	g.WithWriteLock(func() {
		err = g.TakeExtraTurnAfterReplacedMoveForEffect(&ReplacementEvent{Kind: RepEventLife}, uuid.Nil, g.Seats[0].ID)
	})
	if !errors.Is(err, ErrInvalidParam) {
		t.Fatalf("err = %v, want ErrInvalidParam", err)
	}
	g.WithWriteLock(func() { err = g.TakeExtraTurnAfterReplacedMoveForEffect(nil, uuid.Nil, g.Seats[0].ID) })
	if !errors.Is(err, ErrInvalidParam) {
		t.Fatalf("nil event: err = %v, want ErrInvalidParam", err)
	}
}
