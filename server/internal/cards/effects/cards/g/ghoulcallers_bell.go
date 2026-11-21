package cardsg

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Ghoulcaller's Bell — Artifact {1} (Innistrad, #5572):
//
//	"{T}: Each player mills a card."
//
// Vantress Gargoyle's tap ability: each player mills one card, in turn
// order (b02EachPlayerMills).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "36e1485d-3b8c-4c23-9d12-6937fd216b25",
		Name:         "Ghoulcaller's Bell",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "{T}: Each player mills a card.",
			Purpose: game.Purpose{Answers: game.AnswerValue},
			Cost:    game.AbilityCost{Tap: true},
			Effect: func(g *game.Game, item *game.StackItem) error {
				return effects.B02EachPlayerMills(g, item, 1)
			},
		}},
	})
}
