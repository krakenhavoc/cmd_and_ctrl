package cardsch

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Cinderbones — Creature — Elemental Skeleton {2}{B}, 1/1 (Shadowmoor, #5591):
//
//	"Wither
//	 {1}{B}: Regenerate this creature."
//
// Wither rides PrintedKeywords (CR 702.80a); the activated ability makes a
// regeneration shield (CR 701.19a) as River Boa's does.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "8e3c2f44-c4ca-4555-80c6-316e62c338dd",
		Name:            "Cinderbones",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"wither"},
		Activated: []effects.ActivatedAbility{{
			Label:   "{1}{B}: Regenerate this creature.",
			Purpose: game.Purpose{Answers: game.AnswerSurvive},
			Cost:    effects.ManaCost("{1}{B}"),
			Effect: func(g *game.Game, item *game.StackItem) error {
				return effects.Regenerate{Target: item.SourceCardID}.Apply(effects.NewContext(g, item))
			},
		}},
	})
}
