package cardsso

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Spectral Flight — Enchantment — Aura {1}{U} (Innistrad, #5572):
//
//	"Enchant creature
//	 Enchanted creature gets +2/+2 and has flying."
//
// Astral Wingspan's static pair without its draw: a layer-7c +2/+2 and a
// layer-6 flying grant on the enchanted creature.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "d127d1e6-f3a6-4f68-bea5-d7b4123476e5",
		Name:         "Spectral Flight",
		Completeness: effects.CompletenessFull,
		Targets:      effects.EnchantCreature(),
		Purpose:      game.Purpose{Aura: game.AuraBenefit},
		Static: []game.StaticAbility{
			effects.PumpAttached(2, 2),
			effects.GrantToAttached("flying"),
		},
	})
}
