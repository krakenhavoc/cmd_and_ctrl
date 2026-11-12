package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Cloak of Mists — Enchantment — Aura {1}{U} (Urza's Saga):
//
//	"Enchant creature
//	 Enchanted creature can't be blocked."
//
// Aqueous Form without the scry: the attached creature carries the
// can't-be-blocked restriction while the Aura is on it.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "13358b1d-0a35-48ed-a7fb-70359fa3998f",
		Name:         "Cloak of Mists",
		Completeness: CompletenessFull,
		Targets:      EnchantCreature(),
		Purpose:      game.Purpose{Aura: game.AuraBenefit},
		Static: []game.StaticAbility{
			RestrictAttached(game.CantBeBlocked),
		},
	})
}
