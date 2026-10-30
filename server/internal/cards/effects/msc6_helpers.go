package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// msc6_helpers.go — shared bodies for Marvel Super Heroes Commander, batch 6
// (#4455).

// msc6ThisAttacks is "whenever this creature attacks" as an On condition.
func msc6ThisAttacks(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) bool {
	return attackDeclared(ev, source)
}
