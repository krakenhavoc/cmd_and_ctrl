package effectstest

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// runStateChecksViaDraw runs the state-based actions from outside the
// game package: a sandbox draw is a special action that runs them
// (CR 117.5). Seat 1 draws, so it is never the active seat's no-op
// draw-step draw.
func RunStateChecksViaDraw(t *testing.T, g *game.Game) {
	t.Helper()
	if err := g.DrawCard(g.Seats[1].ID); err != nil {
		t.Fatalf("DrawCard: %v", err)
	}
}

// gainAndLoseACounter puts a +1/+1 counter on `id` and takes it off
// again, the add_counter action both ways, then runs the state checks.
func GainAndLoseACounter(t *testing.T, g *game.Game, id uuid.UUID) {
	t.Helper()
	if err := g.AddCounter(id, game.CounterPlusOne, 1); err != nil {
		t.Fatalf("add counter: %v", err)
	}
	RunStateChecksViaDraw(t, g)
	if err := g.AddCounter(id, game.CounterPlusOne, -1); err != nil {
		t.Fatalf("remove counter: %v", err)
	}
	RunStateChecksViaDraw(t, g)
}
