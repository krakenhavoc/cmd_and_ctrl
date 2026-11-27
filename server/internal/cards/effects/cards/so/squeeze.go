package cardsso

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Squeeze — Enchantment {3}{U} (Mercadian Masques, #5619):
//
//	"Sorcery spells cost {3} more to cast."
//
// Thorn of Amethyst's symmetric tax at {3} over sorcery spells.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "e18d8326-c3c3-43c2-ac13-6861c0ef3888",
		Name:         "Squeeze",
		Completeness: effects.CompletenessFull,
		CostModifiers: []game.CostModifier{
			effects.CostsMore(3, "Sorcery spells cost {3} more to cast.", effects.Mmq695SorcerySpell()),
		},
	})
}
