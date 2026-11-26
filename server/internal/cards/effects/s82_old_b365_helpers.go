package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// S82OldB365TappedLands is "the number of tapped lands on the battlefield",
// every player's (Ignition Team).
func S82OldB365TappedLands(g *game.Game) int {
	n := 0
	for _, c := range g.BattlefieldCardsForEffect() {
		if c.IsLand() && c.Tapped {
			n++
		}
	}
	return n
}
