package cardsto

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Trestle Troll — Creature — Troll {1}{B}{G}, 1/4 (Return to Ravnica, #5566):
//
//	"Defender
//	 Reach
//	 {1}{B}{G}: Regenerate this creature."
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "8661d773-3b92-4fcf-8ffc-aced1ff7be8a",
		Name:            "Trestle Troll",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"defender", "reach"},
		Activated: []effects.ActivatedAbility{{
			Label:   "{1}{B}{G}: Regenerate this creature.",
			Purpose: game.Purpose{Answers: game.AnswerSurvive},
			Cost:    effects.ManaCost("{1}{B}{G}"),
			Effect: func(g *game.Game, item *game.StackItem) error {
				return effects.Regenerate{Target: item.SourceCardID}.Apply(effects.NewContext(g, item))
			},
		}},
	})
}
