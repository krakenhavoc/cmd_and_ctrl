package game

import (
	"testing"

	"github.com/google/uuid"
)

// library_to_hand_test.go — #952. The named library-to-hand move: a
// door of its own rather than a bounce that happens to work.

// pushLibraryTop puts a card on top of `owner`'s library and returns
// its ID.
func pushLibraryTop(g *Game, owner *Player, name string) uuid.UUID {
	id := uuid.New()
	owner.Library.PushTop(Card{
		InstanceID: id, Name: name, TypeLine: "Creature — Goblin",
		Owner: owner.ID, Controller: owner.ID,
	})
	return id
}

// TestTakeFromLibraryToHandMovesAndReportsWhatLanded.
func TestTakeFromLibraryToHandMovesAndReportsWhatLanded(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0]
	a := pushLibraryTop(g, me, "Goblin Matron")
	b := pushLibraryTop(g, me, "Goblin Piledriver")

	var landed []uuid.UUID
	g.WithWriteLock(func() {
		if err := g.TakeFromLibraryToHandThenForEffect(me.ID, []uuid.UUID{a, b},
			func(_ *Game, taken []uuid.UUID) error {
				landed = append([]uuid.UUID(nil), taken...)
				return nil
			}); err != nil {
			t.Fatalf("TakeFromLibraryToHandThenForEffect: %v", err)
		}
	})
	if len(landed) != 2 {
		t.Fatalf("landed = %d cards, want 2", len(landed))
	}
	for _, id := range []uuid.UUID{a, b} {
		if !me.Hand.Contains(id) {
			t.Errorf("%s is not in hand", id)
		}
		if me.Library.Contains(id) {
			t.Errorf("%s is still in the library", id)
		}
	}
}

// A card that is not in a library is skipped. This is the whole
// difference between the new door and BounceCardsToHandForEffect,
// which finds a card's zone by scan and would return a battlefield
// permanent to hand under the name of a library take.
func TestTakeFromLibraryToHandSkipsACardThatIsNotInALibrary(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0]
	inLibrary := pushLibraryTop(g, me, "Goblin Matron")
	onBoard := pushCreatureToBattlefield(t, g, me)

	var landed []uuid.UUID
	g.WithWriteLock(func() {
		if err := g.TakeFromLibraryToHandThenForEffect(me.ID, []uuid.UUID{onBoard, inLibrary},
			func(_ *Game, taken []uuid.UUID) error {
				landed = append([]uuid.UUID(nil), taken...)
				return nil
			}); err != nil {
			t.Fatalf("TakeFromLibraryToHandThenForEffect: %v", err)
		}
	})
	if len(landed) != 1 || landed[0] != inLibrary {
		t.Fatalf("landed = %v, want just the library card %s", landed, inLibrary)
	}
	if !g.Battlefield.Contains(onBoard) {
		t.Error("the battlefield permanent was moved — this is a LIBRARY to hand move")
	}
	if me.Hand.Contains(onBoard) {
		t.Error("the battlefield permanent reached a hand")
	}
}

// "Its owner's hand", not the taker's. The player argument is who is
// doing the taking (it is stamped on the events); the destination is
// resolved per card, the way every other route resolves an owner's
// zone.
func TestTakeFromLibraryToHandGoesToTheOwnersHand(t *testing.T) {
	g := newFourPlayerActiveGame(t)
	taker, owner := g.Seats[0], g.Seats[1]
	theirs := pushLibraryTop(g, owner, "Goblin Matron")

	g.WithWriteLock(func() {
		if err := g.TakeFromLibraryToHandThenForEffect(taker.ID, []uuid.UUID{theirs}, nil); err != nil {
			t.Fatalf("TakeFromLibraryToHandThenForEffect: %v", err)
		}
	})
	if !owner.Hand.Contains(theirs) {
		t.Error("the card did not reach its OWNER's hand")
	}
	if taker.Hand.Contains(theirs) {
		t.Error("the card reached the TAKER's hand — a card can only go to its owner's")
	}
}

// The continuation runs after the move, and what it sees is the board
// afterwards: the taken cards gone from the library, the rest still
// there. That ordering is why "put the rest on the bottom" belongs
// inside it.
func TestTakeFromLibraryToHandRunsItsContinuationAfterTheMove(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0]
	taken := pushLibraryTop(g, me, "Goblin Matron")
	left := pushLibraryTop(g, me, "Mountain")

	ran := 0
	g.WithWriteLock(func() {
		if err := g.TakeFromLibraryToHandThenForEffect(me.ID, []uuid.UUID{taken},
			func(g *Game, _ []uuid.UUID) error {
				ran++
				if !me.Hand.Contains(taken) {
					t.Error("the continuation ran before the move landed")
				}
				if !me.Library.Contains(left) {
					t.Error("the continuation cannot see the rest of the library")
				}
				return nil
			}); err != nil {
			t.Fatalf("TakeFromLibraryToHandThenForEffect: %v", err)
		}
	})
	if ran != 1 {
		t.Errorf("the continuation ran %d times, want exactly 1", ran)
	}
}

// An empty ask still runs the continuation, so a card whose look found
// nothing still puts "the rest" where it said.
func TestTakeFromLibraryToHandWithNothingToTakeStillContinues(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0]
	ran := false
	g.WithWriteLock(func() {
		if err := g.TakeFromLibraryToHandThenForEffect(me.ID, nil,
			func(*Game, []uuid.UUID) error { ran = true; return nil }); err != nil {
			t.Fatalf("TakeFromLibraryToHandThenForEffect: %v", err)
		}
	})
	if !ran {
		t.Error("the continuation did not run for an empty take")
	}
}
