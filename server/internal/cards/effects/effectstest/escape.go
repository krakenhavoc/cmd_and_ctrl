package effectstest

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// escape_test.go — S29. The game package pins the mechanic
// (escape_test.go there: the count, the distinctness, "other",
// "your"); this file pins the seven cards against the REAL catalog.
//
// That is the half a stubbed hook cannot check — that each card file
// declared BOTH the zone and the price, that Register accepted the
// pair, and that the offer the client would render is the offer the
// cast path accepts.

// seedGraveyardFodder puts n throwaway cards in the active seat's
// graveyard and returns their IDs. Escape's cost is the only one in
// the catalog that needs a stocked graveyard to be payable at all.
func SeedGraveyardFodder(t *testing.T, g *game.Game, n int) []uuid.UUID {
	t.Helper()
	active := g.Seats[g.Turn.ActiveSeat]
	out := make([]uuid.UUID, 0, n)
	for i := 0; i < n; i++ {
		c := game.NewCard("Fodder", active.ID)
		c.TypeLine = "Instant"
		active.Graveyard.PushTop(c)
		out = append(out, c.InstanceID)
	}
	return out
}
