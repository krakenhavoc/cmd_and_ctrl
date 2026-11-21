package cardsco

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Crystal Ball — Artifact {3}:
//
//	"{1}, {T}: Scry 2."
//
// The repeatable scry. Nothing clever: it is the Scry primitive
// behind an ordinary mana-plus-tap activation, and it is in the
// catalog because a draw deck wants an engine that smooths every
// turn rather than a one-shot that smooths once.
//
// No simplifications.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "bd85fe4d-1d62-416f-ac2d-e287911c84e3",
		Name:         "Crystal Ball",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "{1}, {T}: Scry 2.",
			Purpose: game.Purpose{Answers: game.AnswerValue, Scry: 2},
			Cost:    game.AbilityCost{Mana: "{1}", Tap: true},
			Effect: func(g *game.Game, item *game.StackItem) error {
				return effects.Scry{Player: item.Controller, N: 2}.Apply(effects.NewContext(g, item))
			},
		}},
	})
}
