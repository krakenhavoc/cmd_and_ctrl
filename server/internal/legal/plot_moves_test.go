package legal_test

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/legal"
)

// plot_moves_test.go — #1318: a plotted card (CR 702.170d) is offered
// to the bot exactly when the engine accepts the cast — its owner's
// main phase, stack empty, a later turn — and for nothing.

func TestEnumeratorOffersAPlottedCastOnlyInTheMainPhase(t *testing.T) {
	g := newTable(t)
	advanceTo(t, g, game.StepUpkeep)
	seat := g.Seats[g.Turn.ActiveSeat]
	clearHand(seat)

	c := game.NewCard("Plotted Instant", seat.ID)
	c.TypeLine = "Instant"
	c.ManaCost = "{3}{U}"
	c.Layout = "normal"
	c.Controller = seat.ID
	g.Exile.PushTop(c)
	g.WithWriteLock(func() {
		g.PlotExiledCardForEffect(c.InstanceID, uuid.Nil)
		g.Turn.Number++ // a later turn
	})

	// An instant, and it is the owner's turn — but the upkeep is not
	// the plot window.
	if got := castMovesFor(legal.EnumerateFor(g, seat.ID), c.InstanceID); len(got) != 0 {
		t.Fatalf("a plotted instant was offered in the upkeep: %d moves", len(got))
	}

	advanceTo(t, g, game.StepPrecombatMain)
	moves := legal.EnumerateFor(g, seat.ID)
	if got := castMovesFor(moves, c.InstanceID); len(got) == 0 {
		t.Fatalf("no plotted cast offered in the main phase with no mana: %v", labels(moves))
	}
	dispatchAll(t, g, seat.ID, moves)
}
