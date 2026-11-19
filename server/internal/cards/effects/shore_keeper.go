package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Shore Keeper — Creature — Trilobite {U}, 0/3 (Foundations Jumpstart):
//
//	"{7}{U}, {T}, Sacrifice this creature: Draw three cards."
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "9625773b-1976-49b3-88d2-93a4befb595c",
		Name:         "Shore Keeper",
		Completeness: CompletenessFull,
		Activated: []ActivatedAbility{{
			Label:   "{7}{U}, {T}, Sacrifice this creature: Draw three cards.",
			Purpose: game.Purpose{Answers: game.AnswerValue},
			Cost:    game.AbilityCost{Mana: "{7}{U}", Tap: true, SacrificeSelf: true},
			Effect:  Do(DrawCards{N: 3}),
		}},
	})
}
