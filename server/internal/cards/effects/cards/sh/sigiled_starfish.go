package cardssh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Sigiled Starfish — Creature — Starfish {1}{U}, 0/3 (Foundations
// Jumpstart):
//
//	"{T}: Scry 1."
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "94a4f70a-8aac-412a-99f0-933e3ff87fe0",
		Name:         "Sigiled Starfish",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "{T}: Scry 1.",
			Purpose: game.Purpose{Answers: game.AnswerValue, Scry: 1},
			Cost:    effects.TapCost(),
			Effect: func(g *game.Game, item *game.StackItem) error {
				return effects.Scry{Player: item.Controller, N: 1}.Apply(effects.NewContext(g, item))
			},
		}},
	})
}
