package cardssh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Shackles — Enchantment — Aura {2}{W} (Invasion, S82 older sets batch 661, #5616):
//
//	"Enchant creature
//	 Enchanted creature doesn't untap during its controller's untap step.
//	 {W}: Return this Aura to its owner's hand."
//
// Conviction's return ability with a static on the enchanted creature.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:              "05861ff2-ec85-43d1-b641-c5d2da819c03",
		Name:                  "Shackles",
		Completeness:          effects.CompletenessFull,
		Targets:               effects.EnchantCreature(),
		Purpose:               game.Purpose{Aura: game.AuraHarm},
		UntapStepRestrictions: []game.UntapStepRestriction{effects.EnchantedDoesntUntap()},
		Activated: []effects.ActivatedAbility{{
			Label:   "{W}: Return this Aura to its owner's hand.",
			Purpose: game.Purpose{Answers: game.AnswerSurvive},
			Cost:    effects.ManaCost("{W}"),
			Effect:  effects.Do(effects.Instruction(effects.StdB9ReturnThisToHand)),
		}},
	})
}
