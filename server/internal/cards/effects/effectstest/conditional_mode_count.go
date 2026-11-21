package effectstest

import (
	"testing"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// cmcSeats is the active seat and an opponent, with the table walked
// to the active seat's main phase.
func CmcSeats(t *testing.T, g *game.Game) (*game.Player, *game.Player) {
	t.Helper()
	AdvanceToMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	return me, opp
}

func RedIn(p *game.Player) int {
	n := 0
	for _, tok := range p.ManaPool.Units() {
		if tok.Color == "R" {
			n++
		}
	}
	return n
}
