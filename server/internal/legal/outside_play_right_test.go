package legal_test

import (
	"strings"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/legal"
)

// outside_play_right_test.go — #3955, Wish: a seat whose pool is the card
// database is offered the action that names a card to play, while it holds
// the right and its container is empty; a sideboard pool casts its cards
// directly and is never offered it.

func outsidePlayMoves(moves []legal.Move) []legal.Move {
	var out []legal.Move
	for _, m := range moves {
		if m.Type == legal.TypeOutsidePlay {
			out = append(out, m)
		}
	}
	return out
}

func TestTheDatabasePoolIsOfferedTheNamingActionUnderTheRight(t *testing.T) {
	installLegalCollection(t)
	g := newSideboardTable(t) // no sideboard on either seat
	seat := g.Seats[g.Turn.ActiveSeat]
	clearHand(seat)
	advanceTo(t, g, game.StepPrecombatMain)

	if got := outsidePlayMoves(legal.EnumerateFor(g, seat.ID)); len(got) != 0 {
		t.Fatalf("offered the naming action with no right: %d moves", len(got))
	}
	g.WithWriteLock(func() { g.GrantOutsidePlayRightForEffect(seat.ID, uuid.New(), "Wish") })
	moves := legal.EnumerateFor(g, seat.ID)
	got := outsidePlayMoves(moves)
	if len(got) != 1 || !strings.Contains(got[0].Label, "Wish") || got[0].Kind != legal.KindSpecialAction {
		t.Fatalf("naming action moves: %+v", got)
	}
	dispatchAll(t, g, seat.ID, moves)

	// The other seat holds no right and is offered nothing.
	other := g.Seats[1-g.Turn.ActiveSeat]
	if got := outsidePlayMoves(legal.EnumerateFor(g, other.ID)); len(got) != 0 {
		t.Errorf("a seat with no right was offered the naming action")
	}
}

// Once a card is named (the container holds it) the bot is not offered
// the action again; it casts the card, or waits for the card's timing.
func TestTheNamingActionIsNotOfferedOverAndOver(t *testing.T) {
	installLegalCollection(t)
	g := newSideboardTable(t)
	seat := g.Seats[g.Turn.ActiveSeat]
	clearHand(seat)
	advanceTo(t, g, game.StepPrecombatMain)
	g.WithWriteLock(func() { g.GrantOutsidePlayRightForEffect(seat.ID, uuid.New(), "Wish") })
	if err := g.NameOutsideCardToPlay(seat.ID); err != nil {
		t.Fatal(err)
	}
	pick := g.PendingChoices[len(g.PendingChoices)-1]
	if err := g.ResolveOutsideCollection(pick.ID, seat.ID, []string{"Alpha Sorcery"}); err != nil {
		t.Fatal(err)
	}
	if got := outsidePlayMoves(legal.EnumerateFor(g, seat.ID)); len(got) != 0 {
		t.Errorf("offered the naming action again with a named card waiting: %d moves", len(got))
	}
}

func TestASideboardPoolIsNeverOfferedTheNamingAction(t *testing.T) {
	installLegalCollection(t)
	g, seat, _ := newOutsideTable(t, "Creature — Bear", "{2}")
	g.WithWriteLock(func() { g.GrantOutsidePlayRightForEffect(seat.ID, uuid.New(), "Wish") })
	if got := outsidePlayMoves(legal.EnumerateFor(g, seat.ID)); len(got) != 0 {
		t.Errorf("a sideboard pool was offered the naming action: %d moves", len(got))
	}
}
