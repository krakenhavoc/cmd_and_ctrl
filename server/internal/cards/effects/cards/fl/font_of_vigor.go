package cardsfl

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Font of Vigor — Enchantment {1}{W} (Journey into Nyx, #5558):
//
//	"{2}{W}, Sacrifice this enchantment: You gain 7 life."
//
// A sacrifice-as-cost ability, activated at instant speed.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "422876dd-2801-4d0e-93ad-d9d48178f0d0",
		Name:         "Font of Vigor",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "{2}{W}, Sacrifice this enchantment: You gain 7 life.",
			Purpose: game.Purpose{Answers: game.AnswerValue},
			Cost:    game.AbilityCost{SacrificeSelf: true, Mana: "{2}{W}"},
			Effect:  effects.Do(effects.GainLife{Amount: 7}),
		}},
	})
}
