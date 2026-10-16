package legal_test

import (
	"fmt"
	"strings"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/legal"
)

// outside_cast_test.go — #3955: a card its owner owns outside the game is
// a cast the enumerator offers, only under a permission, and every move
// it offers is one the engine accepts.

// newOutsideTable seats a sideboard card on EVERY seat, so whichever seat
// goes first owns one, with the first seat's precombat main reached.
func newOutsideTable(t *testing.T, typeLine, cost string) (*game.Game, *game.Player, game.Card) {
	t.Helper()
	g := game.NewGame()
	for i := 0; i < 2; i++ {
		deck := []game.Card{game.NewCommander(fmt.Sprintf("Commander %d", i), uuid.Nil)}
		for j := 0; j < 20; j++ {
			deck = append(deck, game.NewCard(fmt.Sprintf("Filler %d-%d", i, j), uuid.Nil))
		}
		c := game.NewCard(fmt.Sprintf("Outside Card %d", i), uuid.Nil)
		c.TypeLine = typeLine
		c.ManaCost = cost
		c.IsSideboard = true
		deck = append(deck, c)
		if _, err := g.AddPlayer(fmt.Sprintf("Seat%d", i), deck); err != nil {
			t.Fatalf("AddPlayer: %v", err)
		}
	}
	if err := g.StartWithSource(nil); err != nil {
		t.Fatalf("Start: %v", err)
	}
	for _, p := range g.Seats {
		if err := g.KeepHand(p.ID); err != nil {
			t.Fatalf("KeepHand: %v", err)
		}
	}
	seat := g.Seats[g.Turn.ActiveSeat]
	clearHand(seat)
	advanceTo(t, g, game.StepPrecombatMain)
	return g, seat, seat.SideboardCards()[0]
}

func TestEnumeratorOffersAFreeCastFromOutsideOnlyUnderAPermission(t *testing.T) {
	g, seat, card := newOutsideTable(t, "Creature — Eldrazi", "{10}")

	if got := castMovesFor(legal.EnumerateFor(g, seat.ID), card.InstanceID); len(got) != 0 {
		t.Fatalf("a sideboard card with no permission was offered %d casts", len(got))
	}

	g.WithWriteLock(func() { g.FreeCastFromOutsideForEffect(seat.ID, uuid.New(), card.InstanceID, "Test Source") })
	moves := legal.EnumerateFor(g, seat.ID)
	got := castMovesFor(moves, card.InstanceID)
	if len(got) == 0 {
		t.Fatalf("no free cast from outside offered: %v", labels(moves))
	}
	if !strings.Contains(got[0].Label, "from outside the game") {
		t.Errorf("label %q does not say where the card is cast from", got[0].Label)
	}
	dispatchAll(t, g, seat.ID, moves)
}

func TestEnumeratorOffersAPlayFromOutsideAtItsCost(t *testing.T) {
	g, seat, card := newOutsideTable(t, "Creature — Bear", "{2}")
	g.WithWriteLock(func() { g.GrantOutsidePlayRightForEffect(seat.ID, uuid.New(), "Test Wish") })

	// A play pays the card's own cost: unfunded, the bot is offered
	// nothing (a free cast would have been offered); funded, it is.
	if got := castMovesFor(legal.EnumerateFor(g, seat.ID), card.InstanceID); len(got) != 0 {
		t.Fatalf("an unaffordable play from outside was offered: %d moves", len(got))
	}
	mana(g, seat, 2)
	moves := legal.EnumerateFor(g, seat.ID)
	if got := castMovesFor(moves, card.InstanceID); len(got) == 0 {
		t.Fatalf("no play from outside offered: %v", labels(moves))
	}
	dispatchAll(t, g, seat.ID, moves)
}

func TestEnumeratorOffersALandPlayFromOutside(t *testing.T) {
	g, seat, card := newOutsideTable(t, "Basic Land — Forest", "")
	g.WithWriteLock(func() { g.GrantOutsidePlayRightForEffect(seat.ID, uuid.New(), "Test Wish") })
	moves := legal.EnumerateFor(g, seat.ID)
	found := false
	for _, m := range moves {
		if m.Source == card.InstanceID {
			found = true
		}
	}
	if !found {
		t.Fatalf("no land play from outside offered: %v", labels(moves))
	}
	dispatchAll(t, g, seat.ID, moves)

	// A free (cast-only) grant offers no land play.
	g2, seat2, card2 := newOutsideTable(t, "Basic Land — Forest", "")
	g2.WithWriteLock(func() { g2.FreeCastFromOutsideForEffect(seat2.ID, uuid.New(), card2.InstanceID, "Test Source") })
	for _, m := range legal.EnumerateFor(g2, seat2.ID) {
		if m.Source == card2.InstanceID {
			t.Errorf("a cast-only grant offered the land: %q", m.Label)
		}
	}
}
