package cardslu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

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
	effects.Register(effects.Spec{
		OracleID:     "ea8c4a7d-f7fb-4620-b3a4-ae8280dd60d9",
		Name:         "Magefire Wings",
		Completeness: effects.CompletenessFull,
		Targets:      effects.EnchantCreature(),
		Static: []game.StaticAbility{
			effects.PumpAttached(2, 0),
			effects.GrantToAttached("flying"),
		},
	})
}
