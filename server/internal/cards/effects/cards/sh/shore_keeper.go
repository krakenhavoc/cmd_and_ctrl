package cardssh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Shore Keeper — Creature — Trilobite {U}, 0/3 (Foundations Jumpstart):
//
//	"{7}{U}, {T}, Sacrifice this creature: Draw three cards."
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "9625773b-1976-49b3-88d2-93a4befb595c",
		Name:         "Shore Keeper",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "{7}{U}, {T}, Sacrifice this creature: Draw three cards.",
			Purpose: game.Purpose{Answers: game.AnswerValue},
			Cost:    game.AbilityCost{Mana: "{7}{U}", Tap: true, SacrificeSelf: true},
			Effect:  effects.Do(effects.DrawCards{N: 3}),
		}},
	})
}
