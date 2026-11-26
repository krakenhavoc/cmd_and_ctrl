package cardshu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Invisibility — Enchantment — Aura {U}{U} (Magic 2015, #5556):
//
//	"Enchant creature
//	 Enchanted creature can't be blocked except by Walls."
//
// Evil Eye of Orms-by-Gore's block rule, scoped to the creature the Aura is
// attached to (OnAttached) rather than to its own source: it ends the moment
// the Aura does.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "de26b0c6-dfb7-45a8-9d7f-f8d45522d675",
		Name:         "Invisibility",
		Completeness: effects.CompletenessFull,
		Targets:      effects.EnchantCreature(),
		Purpose:      game.Purpose{Aura: game.AuraBenefit},
		BlockRules: []game.BlockRule{
			effects.CantBeBlockedExceptBy(effects.OnAttached(), effects.OfCreatureType("Wall"), "Walls"),
		},
	})
}
