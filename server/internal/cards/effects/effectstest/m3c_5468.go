package effectstest

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// m3c_5468.go — Modern Horizons 3 Commander batch 1 (#5468) helpers shared across test packages.

// M3cb1Opponent is the seat after the active one.
func M3cb1Opponent(g *game.Game) *game.Player {
	return g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
}
