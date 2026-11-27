package cardsw

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Whip Silk — Enchantment — Aura {G} (Invasion, S82 older sets batch 661, #5616):
//
//	"Enchant creature
//	 Enchanted creature has reach.
//	 {G}: Return this Aura to its owner's hand."
//
// Conviction's return ability with a static on the enchanted creature.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "ce2c420b-e569-4715-b65b-4a097700799d",
		Name:         "Whip Silk",
		Completeness: effects.CompletenessFull,
		Targets:      effects.EnchantCreature(),
		Purpose:      game.Purpose{Aura: game.AuraBenefit},
		Static:       []game.StaticAbility{effects.GrantToAttached("reach")},
		Activated: []effects.ActivatedAbility{{
			Label:   "{G}: Return this Aura to its owner's hand.",
			Purpose: game.Purpose{Answers: game.AnswerSurvive},
			Cost:    effects.ManaCost("{G}"),
			Effect:  effects.Do(effects.Instruction(effects.StdB9ReturnThisToHand)),
		}},
	})
}
