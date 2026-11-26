package cardss

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Selesnya Sentry — Creature — Elephant Soldier {2}{W}, 3/2 (Return to Ravnica,
// #5566):
//
//	"{5}{G}: Regenerate this creature."
//
// The shield is for this permanent only while it is still the one that
// activated the ability.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "29e8d973-0233-446e-b587-1c7508352642",
		Name:         "Selesnya Sentry",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "{5}{G}: Regenerate this creature.",
			Purpose: game.Purpose{Answers: game.AnswerSurvive},
			Cost:    effects.ManaCost("{5}{G}"),
			Effect: func(g *game.Game, item *game.StackItem) error {
				return effects.Regenerate{Target: item.SourceCardID}.Apply(effects.NewContext(g, item))
			},
		}},
	})
}
