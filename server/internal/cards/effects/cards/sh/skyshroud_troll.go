package cardssh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Skyshroud Troll — Creature — Troll Giant {2}{G}{G}, 3/3 (Tempest, S82 OLD
// b752, #5628):
//
//	"{1}{G}: Regenerate this creature."
//
// Albino Troll's ability: Regenerate on the ability's own source (CR 701.19).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "2e6d3120-d41d-4e48-b33f-0f73c876fe69",
		Name:         "Skyshroud Troll",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "{1}{G}: Regenerate this creature.",
			Purpose: game.Purpose{Answers: game.AnswerSurvive},
			Cost:    effects.ManaCost("{1}{G}"),
			Effect: func(g *game.Game, item *game.StackItem) error {
				return effects.Regenerate{Target: item.SourceCardID}.Apply(effects.NewContext(g, item))
			},
		}},
	})
}
