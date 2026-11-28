package cardsw

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Wings of Aesthir — Enchantment — Aura {W}{U} (Ice Age, S82 OLD b828, #5636):
//
//	"Enchant creature
//	 Enchanted creature gets +1/+0 and has flying and first strike."
//
// Arcane Flight's shape: PumpAttached and a keyword grant.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "06413d87-d119-4c04-93d5-5ced7ad4a858",
		Name:         "Wings of Aesthir",
		Completeness: effects.CompletenessFull,
		Targets:      effects.EnchantCreature(),
		Purpose:      game.Purpose{Aura: game.AuraBenefit},
		Static:       []game.StaticAbility{effects.PumpAttached(1, 0), effects.GrantToAttached("flying", "first strike")},
	})
}
