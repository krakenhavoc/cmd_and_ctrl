package effectstest

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

func PhasedOutInCatalogGame(g *game.Game, id uuid.UUID) bool {
	found := false
	g.ReadSnapshot(func() { found = g.IsPhasedOutForEffect(id) })
	return found
}

// passTurnsTo walks the table round to `seat`'s own turn, so CR 502.1
// runs inside the real untap step rather than through a test hook.
// `seat` is an index into g.Seats.
func PassTurnsTo(t *testing.T, g *game.Game, seat int) {
	t.Helper()
	for i := 0; i < len(g.Seats)+1; i++ {
		if g.Turn.ActiveSeat == seat && i > 0 {
			return
		}
		if err := g.PassTurn(); err != nil {
			t.Fatalf("PassTurn: %v", err)
		}
		if g.Turn.ActiveSeat == seat {
			return
		}
	}
	t.Fatalf("never reached seat %d's turn", seat)
}

// settleStateBasedActions walks the cursor one step, which is the
// public path that runs the SBA pass — and with it the CR 611.2b
// sweep that releases a "phases out until ~ leaves the battlefield".
func SettleStateBasedActions(t *testing.T, g *game.Game) {
	t.Helper()
	if _, err := g.AdvanceStep(); err != nil {
		t.Fatalf("AdvanceStep: %v", err)
	}
}
