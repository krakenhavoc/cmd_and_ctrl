package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// A Tale for the Ages — Enchantment {1}{W} (Wilds of Eldraine, #4471):
//
//	"Enchanted creatures you control get +2/+2."
//
// Syr Armont, the Redeemer's anthem at +2/+2: a layer 7c modify on every
// creature its controller controls that has an Aura attached, whoever
// controls the Aura, re-read on each layer pass so it starts and stops
// with the Aura.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "682a0770-8fe0-4142-aa0c-b60af3628a33",
		Name:         "A Tale for the Ages",
		Completeness: CompletenessFull,
		Static: []game.StaticAbility{{
			Label:    "Enchanted creatures you control get +2/+2.",
			Layer:    game.Layer7PT,
			SubLayer: game.SubLayer7C_Modify,
			AppliesTo: func(target *game.Card, g *game.Game, source *game.Card) bool {
				return target.IsCreature() && target.Controller == source.Controller && isEnchantedByAura(g, target.InstanceID)
			},
			Apply: func(c *game.Characteristic, _ *game.Card, _ *game.Game, _ *game.Card) {
				c.Power += 2
				c.Toughness += 2
			},
		}},
	})
}
