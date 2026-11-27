package cardsso

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Soul Channeling — Enchantment — Aura {2}{B} (Mercadian Masques, S82 OLD batch 685, #5619):
//
//	"Enchant creature
//	 Pay 2 life: Regenerate enchanted creature."
//
// Strands of Undeath's regenerate-the-host with a life payment as the cost (paid by the Aura's controller), so the shield goes on whatever the Aura is attached to as it resolves.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "7041b3d5-c61a-4b94-811c-c5eb9f1ba664",
		Name:         "Soul Channeling",
		Completeness: effects.CompletenessFull,
		Targets:      effects.EnchantCreature(),
		Purpose:      game.Purpose{Aura: game.AuraBenefit},
		Activated: []effects.ActivatedAbility{{
			Label:   "Pay 2 life: Regenerate enchanted creature.",
			Purpose: game.Purpose{Answers: game.AnswerSurvive},
			Cost:    effects.PayLife(2),
			Effect:  effects.RegenerateAttachedHost,
		}},
	})
}
