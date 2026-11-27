package cardsx

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Zephid's Embrace — Enchantment — Aura {2}{U}{U} (Urza's Saga, #5624):
//
//	"Enchant creature
//	 Enchanted creature gets +2/+2 and has flying and shroud."
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "2f8b07ef-9d00-4eb6-a395-b5033aa3f80e",
		Name:         "Zephid's Embrace",
		Completeness: effects.CompletenessFull,
		Targets:      effects.EnchantCreature(),
		Purpose:      game.Purpose{Aura: game.AuraBenefit},
		Static:       []game.StaticAbility{effects.PumpAttached(2, 2), effects.GrantToAttached("flying", "shroud")},
	})
}
