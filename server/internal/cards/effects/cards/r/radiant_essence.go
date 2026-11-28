package cardsr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Radiant Essence — Creature — Spirit {1}{G}{W}, 2/3 (Mirage, #5633):
//
//	"This creature gets +1/+2 as long as an opponent controls a black permanent."
//
// A layer 7c effect (CR 613.4c) that follows the board: a black permanent coming
// or going moves it at once, through the layers, so a colour-changed permanent
// counts as what it is now.
//
// No simplification.
func init() {
	opponentHasBlack := effects.OpponentControlsAtLeast(1, func(c game.Card) bool { return c.HasColorWord("B") })
	effects.Register(effects.Spec{
		OracleID:     "48e2f2c3-9d99-4535-915f-551fbb59ade5",
		Name:         "Radiant Essence",
		Completeness: effects.CompletenessFull,
		Static: []game.StaticAbility{{
			Label:    "This creature gets +1/+2 as long as an opponent controls a black permanent.",
			Layer:    game.Layer7PT,
			SubLayer: game.SubLayer7C_Modify,
			AppliesTo: func(target *game.Card, g *game.Game, source *game.Card) bool {
				return target.InstanceID == source.InstanceID && opponentHasBlack(g, source.Controller, source.InstanceID)
			},
			Apply: func(c *game.Characteristic, _ *game.Card, _ *game.Game, _ *game.Card) {
				c.Power++
				c.Toughness += 2
			},
		}},
	})
}
