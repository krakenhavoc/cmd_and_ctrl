package cardsd

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Day of Destiny — Legendary Enchantment {3}{W}:
//
//	"Legendary creatures you control get +2/+2."
//
// A layer 7c modification (CR 613.4c), so it stacks with other pumps and
// follows a creature that becomes legendary.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "64247fcb-0e1b-48c9-bdac-769d7edd8d1d",
		Name:         "Day of Destiny",
		Completeness: effects.CompletenessFull,
		Static: []game.StaticAbility{{
			Label:    "Legendary creatures you control get +2/+2.",
			Layer:    game.Layer7PT,
			SubLayer: game.SubLayer7C_Modify,
			AppliesTo: func(target *game.Card, _ *game.Game, source *game.Card) bool {
				return target.Controller == source.Controller && target.IsCreature() && target.IsLegendary()
			},
			Apply: func(c *game.Characteristic, _ *game.Card, _ *game.Game, _ *game.Card) {
				c.Power += 2
				c.Toughness += 2
			},
		}},
	})
}
