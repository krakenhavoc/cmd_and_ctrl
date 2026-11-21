package cardsap

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Arcane Flight — Enchantment — Aura {U} (Dominaria, #5530):
//
//	"Enchant creature
//	 Enchanted creature gets +1/+1 and has flying."
//
// Wings of Hope's shape: +1/+1 and flying on the enchanted creature.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "40eee955-bec6-455a-addf-68cdb84c76b9",
		Name:         "Arcane Flight",
		Completeness: effects.CompletenessFull,
		Targets:      effects.EnchantCreature(),
		Purpose:      game.Purpose{Aura: game.AuraBenefit},
		Static:       []game.StaticAbility{effects.PumpAttached(1, 1), effects.GrantToAttached("flying")},
	})
}
