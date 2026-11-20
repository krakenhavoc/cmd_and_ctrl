package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	Register(Spec{
		OracleID:     "d127d1e6-f3a6-4f68-bea5-d7b4123476e5",
		Name:         "Spectral Flight",
		Completeness: CompletenessFull,
		Targets:      EnchantCreature(),
		Purpose:      game.Purpose{Aura: game.AuraBenefit},
		Static: []game.StaticAbility{
			PumpAttached(2, 2),
			GrantToAttached("flying"),
		},
	})
}
