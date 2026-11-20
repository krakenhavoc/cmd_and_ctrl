package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Frenzied Rage — Enchantment — Aura {1}{R} (Dominaria, #5530):
//
//	"Enchant creature
//	 Enchanted creature gets +2/+1 and has menace. (It can't be blocked except by two or more creatures.)"
//
// Untamed Hunger's Aura: +2/+1 and menace on the enchanted creature.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "1a04c31c-f5e4-43ff-a5b5-207426b8dca6",
		Name:         "Frenzied Rage",
		Completeness: CompletenessFull,
		Targets:      EnchantCreature(),
		Purpose:      game.Purpose{Aura: game.AuraBenefit},
		Static:       []game.StaticAbility{PumpAttached(2, 1), GrantToAttached("menace")},
	})
}
