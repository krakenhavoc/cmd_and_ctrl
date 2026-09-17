package legal_test

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/legal"
)

// discard_choice_test.go — #651. An effect's discard is a
// PendingChoiceChooseCards over the discarder's own hand, so the
// enumerator already knows it; these tests are the promise, because
// the cost of being wrong is a bot seat with an empty move list asleep
// on a live table (#544, #499).
//
// The bot's old route to a discard was cleanupDiscardMoves, which
// reads Game.DiscardPending and labels every move "Discard to hand
// size". That labelled a Mind Rot as a cleanup discard; now the map
// only ever holds CR 514.1's count, so the label is true again.

func TestEffectDiscardIsEnumeratedForTheDiscardingSeat(t *testing.T) {
	g := newTable(t)
	me := g.Seats[0]
	hand := me.Hand.Size()
	if hand < 3 {
		t.Fatalf("setup: opening hand is %d, need at least 3", hand)
	}
	g.WithWriteLock(func() {
		g.QueueDiscardChoiceForEffect(game.DiscardPrompt{
			Player:   me.ID,
			N:        1,
			Question: "Mind Rot — discard a card",
		})
	})

	moves := legal.EnumerateFor(g, me.ID)
	if len(moves) != hand {
		t.Fatalf("enumerated %d answers, want one per card in hand (%d): %v",
			len(moves), hand, labels(moves))
	}
	for _, m := range moves {
		if m.Type != legal.TypeResolveChoice {
			t.Fatalf("a seat owing a discard was offered %q as well", m.Label)
		}
	}
	if !hasLabel(moves, "Mind Rot — discard a card: choose") {
		t.Errorf("the card's own words head the move label: %v", labels(moves))
	}
	// Every answer the enumerator offers is one the engine accepts.
	dispatchAll(t, g, me.ID, moves)
}

func TestEffectDiscardOffersNothingToTheOtherSeats(t *testing.T) {
	g := newTable(t)
	me, them := g.Seats[0], g.Seats[1]
	g.WithWriteLock(func() {
		g.QueueDiscardChoiceForEffect(game.DiscardPrompt{Player: me.ID, N: 1})
	})
	// legal.anyChoiceOpen: while ANY choice is open no other seat is
	// offered anything, because the engine refuses to move the table.
	if moves := legal.EnumerateFor(g, them.ID); len(moves) != 0 {
		t.Errorf("another seat was offered %v while a discard was owed", labels(moves))
	}
}

// TestUpToDiscardOffersTheEmptyAnswer — "discard up to two cards"
// keeps "choose nothing", the answer the engine can never refuse.
func TestUpToDiscardOffersTheEmptyAnswer(t *testing.T) {
	g := newTable(t)
	me := g.Seats[0]
	g.WithWriteLock(func() {
		g.QueueDiscardChoiceForEffect(game.DiscardPrompt{
			Player:   me.ID,
			N:        2,
			UpTo:     true,
			Question: "Discard up to two",
		})
	})
	moves := legal.EnumerateFor(g, me.ID)
	if len(moves) == 0 {
		t.Fatal("a seat owing an up-to discard was offered nothing")
	}
	if moves[0].Label != "Discard up to two: choose nothing" {
		t.Errorf("first answer is %q, want the empty one", moves[0].Label)
	}
	dispatchAll(t, g, me.ID, moves)
}

// TestCleanupDiscardStillRidesItsOwnMap — the other obligation, which
// #651 deliberately left where it was: a CR 514.1 hand-size discard is
// enumerated from Game.DiscardPending as a discard_selection move.
func TestCleanupDiscardStillRidesItsOwnMap(t *testing.T) {
	g := newTable(t)
	me := g.Seats[0]
	g.WithWriteLock(func() { g.DiscardPending = map[uuid.UUID]int{me.ID: 1} })

	moves := legal.EnumerateFor(g, me.ID)
	if len(moves) == 0 {
		t.Fatal("a seat owing a cleanup discard was offered nothing")
	}
	if !hasLabel(moves, "Discard to hand size:") {
		t.Errorf("the cleanup discard keeps its own label: %v", labels(moves))
	}
	dispatchAll(t, g, me.ID, moves)
}
