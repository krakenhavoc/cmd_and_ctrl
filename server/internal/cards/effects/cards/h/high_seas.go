package cardsh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// High Seas — Enchantment {2}{U} (Mercadian Masques, #5619):
//
//	"Red creature spells and green creature spells cost {1} more to cast."
//
// Thorn of Amethyst's symmetric tax over a creature spell that is red, green
// or both (charged {1}, not {2}, for a red-green one). Every player pays it,
// High Seas' controller included.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "9100f213-f4c1-4b15-94f2-e280579894c0",
		Name:         "High Seas",
		Completeness: effects.CompletenessFull,
		CostModifiers: []game.CostModifier{
			effects.CostsMore(1, "Red creature spells and green creature spells cost {1} more to cast.", effects.Mmq695RedOrGreenCreatureSpell()),
		},
	})
}
