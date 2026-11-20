package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Magefire Wings — Enchantment — Aura {U}{R} (Alara Reborn, #5586):
//
//	"Enchant creature
//	 Enchanted creature gets +2/+0 and has flying."
//
// Future Flight's statics without its draw: a layer 7c +2/+0 and a layer 6
// flying grant on the enchanted creature.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "ea8c4a7d-f7fb-4620-b3a4-ae8280dd60d9",
		Name:         "Magefire Wings",
		Completeness: CompletenessFull,
		Targets:      EnchantCreature(),
		Static: []game.StaticAbility{
			PumpAttached(2, 0),
			GrantToAttached("flying"),
		},
	})
}
