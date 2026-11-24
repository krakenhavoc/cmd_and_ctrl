package cardse

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Enchanted Evening — Enchantment {3}{W/U}{W/U} (Secret Lair Drop, #5462):
//
//	"All permanents are enchantments in addition to their other types."
//
// Mycosynth Lattice's first line with Enchantment: one layer 4 static over every
// permanent on the battlefield, idempotent so a permanent that is already an
// enchantment is unchanged. Cards in other zones and spells are not
// permanents and stay as they are.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "c56ea5ce-c8de-46c2-a8dd-bdd41467a28c",
		Name:         "Enchanted Evening",
		Completeness: effects.CompletenessFull,
		Static: []game.StaticAbility{{
			Layer: game.Layer4Type,
			Label: "All permanents are enchantments in addition to their other types.",
			AppliesTo: func(_ *game.Card, _ *game.Game, _ *game.Card) bool {
				return true
			},
			Apply: effects.AddTheEnchantmentTypeOnce,
		}},
	})
}
