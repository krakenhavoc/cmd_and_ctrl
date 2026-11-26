package cardsco

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Conviction — {1}{W} Enchantment — Aura (Aether Revolt, #5536):
//
//	"Enchant creature
//	 Enchanted creature gets +1/+3.
//	 {W}: Return this Aura to its owner's hand."
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "4374b6b1-e691-4823-802d-8de8530aa461",
		Name:         "Conviction",
		Completeness: effects.CompletenessFull,
		Targets:      effects.EnchantCreature(),
		Purpose:      game.Purpose{Aura: game.AuraBenefit},
		Static:       []game.StaticAbility{effects.PumpAttached(1, 3)},
		Activated: []effects.ActivatedAbility{{
			Label:   "{W}: Return this Aura to its owner's hand.",
			Purpose: game.Purpose{Answers: game.AnswerSurvive},
			Cost:    effects.ManaCost("{W}"),
			Effect:  effects.Do(effects.Instruction(effects.StdB9ReturnThisToHand)),
		}},
	})
}
