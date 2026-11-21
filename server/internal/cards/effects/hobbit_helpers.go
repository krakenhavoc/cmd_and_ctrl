package effects

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// hobbitFerociousAttack is the trigger condition of a Hobbit "Ferocious
// — Whenever this creature attacks while you control a creature with
// power 4 or greater": this creature was declared as an attacker, and
// its controller controls a creature with power 4 or greater when that
// happens. It is the Nighthowl Pursuer / Ravening Warg / Wilderland
// Scrounger window; the power is read with the layers applied.
func HobbitFerociousAttack(ev game.Event, source *game.Card, lki game.Characteristic, g *game.Game) bool {
	return ThisAttacked(ev, source, lki, g) && B24ControlsCreatureWithPowerAtLeast(g, source.Controller, 4)
}
