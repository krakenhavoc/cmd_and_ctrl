package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Sigiled Starfish — Creature — Starfish {1}{U}, 0/3 (Foundations
// Jumpstart):
//
//	"{T}: Scry 1."
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "94a4f70a-8aac-412a-99f0-933e3ff87fe0",
		Name:         "Sigiled Starfish",
		Completeness: CompletenessFull,
		Activated: []ActivatedAbility{{
			Label:   "{T}: Scry 1.",
			Purpose: game.Purpose{Answers: game.AnswerValue, Scry: 1},
			Cost:    TapCost(),
			Effect: func(g *game.Game, item *game.StackItem) error {
				return Scry{Player: item.Controller, N: 1}.Apply(NewContext(g, item))
			},
		}},
	})
}
