package cardsd

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Death's Shadow — Creature — Avatar {B}, 13/13 (Mystery Booster 2):
//
//	"This creature gets -X/-X, where X is your life total."
//
// A layer 7c modification that reads the controller's life as the layers
// run, so it follows every life change. At 13 life it is a 0/0 and dies.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "e08722b3-0f79-4c82-b298-603e04a37857",
		Name:         "Death's Shadow",
		Completeness: effects.CompletenessFull,
		Static: []game.StaticAbility{{
			Label:              "This creature gets -X/-X, where X is your life total.",
			Layer:              game.Layer7PT,
			SubLayer:           game.SubLayer7C_Modify,
			DependsOnLifeTotal: true,
			AppliesTo:          effects.SelfOnly,
			Apply: func(c *game.Characteristic, _ *game.Card, g *game.Game, source *game.Card) {
				x := effects.YourLifeTotal(g, source)
				c.Power -= x
				c.Toughness -= x
			},
		}},
	})
}
