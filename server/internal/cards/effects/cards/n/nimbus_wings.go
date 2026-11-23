package cardsn

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Nimbus Wings — Enchantment — Aura {1}{W} (Zendikar, #5583):
//
//	"Enchant creature
//	 Enchanted creature gets +1/+2 and has flying."
//
// Enchant creature is the shared EnchantCreature clause; the pump and
// the keyword are the shared attachment statics.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "0f35e73b-6a38-4185-a7d3-d237d67ba1cd",
		Name:         "Nimbus Wings",
		Completeness: effects.CompletenessFull,
		Targets:      effects.EnchantCreature(),
		Purpose:      game.Purpose{Aura: game.AuraBenefit},
		Static: []game.StaticAbility{
			effects.PumpAttached(1, 2),
			effects.GrantToAttached("flying"),
		},
	})
}
