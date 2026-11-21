package effectstest

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

func JmpB7Seats(g *game.Game) (me, opp *game.Player) {
	return g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
}
