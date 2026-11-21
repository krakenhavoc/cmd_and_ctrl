package cardsdi

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Diabolic Machine — Artifact Creature — Construct {7}, 4/4
// (The Dark, #5638):
//
//	"{3}: Regenerate this creature."
//
// Albino Troll's regeneration row with a generic {3} cost: Regenerate
// on the ability's own source.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "afd9b232-3f46-40c5-b1cf-2a7e8c955eba",
		Name:         "Diabolic Machine",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "{3}: Regenerate this creature.",
			Purpose: game.Purpose{Answers: game.AnswerSurvive},
			Cost:    effects.ManaCost("{3}"),
			Effect: func(g *game.Game, item *game.StackItem) error {
				return effects.Regenerate{Target: item.SourceCardID}.Apply(effects.NewContext(g, item))
			},
		}},
	})
}
