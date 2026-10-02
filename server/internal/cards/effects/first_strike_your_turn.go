package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// firstStrikeDuringYourTurn is "This creature has first strike during
// your turn" (Razorkin Needlehead, Radha, Heart of Keld): a layer 6
// keyword grant to the source itself whose predicate reads the turn,
// re-answered on every recompute. The layer engine invalidates on turn
// change, so the keyword blinks off on an opponent's turn.
func firstStrikeDuringYourTurn() game.StaticAbility {
	return game.StaticAbility{
		Layer: game.Layer6Ability,
		AppliesTo: func(target *game.Card, g *game.Game, source *game.Card) bool {
			return target.InstanceID == source.InstanceID && isActivePlayer(g, source.Controller)
		},
		Apply: func(c *game.Characteristic, _ *game.Card, _ *game.Game, _ *game.Card) {
			for _, a := range c.Abilities {
				if a == "first strike" {
					return
				}
			}
			c.Abilities = append(c.Abilities, "first strike")
		},
	}
}
