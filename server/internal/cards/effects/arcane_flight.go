package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Arcane Flight — Enchantment — Aura {U} (Dominaria, #5530):
//
//	"Enchant creature
//	 Enchanted creature gets +1/+1 and has flying."
//
// Wings of Hope's shape: +1/+1 and flying on the enchanted creature.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "40eee955-bec6-455a-addf-68cdb84c76b9",
		Name:         "Arcane Flight",
		Completeness: CompletenessFull,
		Targets:      EnchantCreature(),
		Purpose:      game.Purpose{Aura: game.AuraBenefit},
		Static:       []game.StaticAbility{PumpAttached(1, 1), GrantToAttached("flying")},
	})
}
