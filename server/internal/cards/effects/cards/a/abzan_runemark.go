package cardsa

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Abzan Runemark — Enchantment — Aura {2}{W} (Fate Reforged,
// #5552):
//
//	"Enchant creature
//	 Enchanted creature gets +2/+2.
//	 Enchanted creature has vigilance as long as you control a black or
//	 green permanent."
//
// Enchant creature is the Aura's target (EnchantCreature). The Runemark
// statics (frfRunemark): +2/+2 in layer 7c always, and vigilance in layer 6
// while you control a black or green permanent, re-read on every recompute.
// "You" is the Aura's controller.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "0a9276fc-387c-4bde-a6d6-b0dd7abee94b",
		Name:         "Abzan Runemark",
		Completeness: effects.CompletenessFull,
		Targets:      effects.EnchantCreature(),
		Static:       effects.FrfRunemark("vigilance", "B", "G"),
	})
}
