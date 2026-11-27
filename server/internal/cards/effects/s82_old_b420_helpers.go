package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// S82OldB420MayLoot is "you may draw a card. If you do, discard a card" as an
// effect body (Izzet Keyrune's combat-damage trigger): drawing is the "may",
// and the discard follows only if the player chose to draw (CelestusAnswer).
func S82OldB420MayLoot(name string) Effect {
	return func(g *game.Game, item *game.StackItem) error {
		return MayChoice{
			Question: name + " — draw a card, then discard a card?",
			Draws:    1,
			YesLabel: "Draw",
			NoLabel:  "Decline",
			Then:     CelestusAnswer,
		}.Apply(NewContext(g, item))
	}
}
