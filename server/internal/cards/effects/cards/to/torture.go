package cardsto

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Torture — Enchantment — Aura {B} (Shadowmoor, S82 older sets):
//
//	"Enchant creature
//	 {1}{B}: Put a -1/-1 counter on enchanted creature."
//
// Primal Cocoon's counter-on-enchanted-creature body as an activated ability of
// the Aura, for the Aura's controller, whoever controls the creature.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "389b6d77-47ab-43a3-99d1-dcc762939cc4",
		Name:         "Torture",
		Completeness: effects.CompletenessFull,
		Targets:      effects.EnchantCreature(),
		Purpose:      game.Purpose{Aura: game.AuraHarm},
		Activated: []effects.ActivatedAbility{{
			Label:   "{1}{B}: Put a -1/-1 counter on enchanted creature.",
			Purpose: game.Purpose{Answers: game.AnswerRemove},
			Cost:    effects.ManaCost("{1}{B}"),
			Effect:  effects.MinusOneCounterOnEnchantedCreature,
		}},
	})
}
