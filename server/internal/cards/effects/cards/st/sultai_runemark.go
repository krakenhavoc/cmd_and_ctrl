package cardsst

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Sultai Runemark — Enchantment — Aura {2}{B} (Fate Reforged,
// #5552):
//
//	"Enchant creature
//	 Enchanted creature gets +2/+2.
//	 Enchanted creature has deathtouch as long as you control a green or
//	 blue permanent. (Any amount of damage it deals to a creature is
//	 enough to destroy it.)"
//
// Enchant creature is the Aura's target (EnchantCreature). The Runemark
// statics (frfRunemark): +2/+2 in layer 7c always, and deathtouch in layer 6
// while you control a green or blue permanent, re-read on every recompute.
// "You" is the Aura's controller.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "59410454-93ff-41f0-ab62-cdaf156163d0",
		Name:         "Sultai Runemark",
		Completeness: effects.CompletenessFull,
		Targets:      effects.EnchantCreature(),
		Static:       effects.FrfRunemark("deathtouch", "G", "U"),
	})
}
