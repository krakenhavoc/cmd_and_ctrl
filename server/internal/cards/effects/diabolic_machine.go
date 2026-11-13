package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	Register(Spec{
		OracleID:     "afd9b232-3f46-40c5-b1cf-2a7e8c955eba",
		Name:         "Diabolic Machine",
		Completeness: CompletenessFull,
		Activated: []ActivatedAbility{{
			Label:   "{3}: Regenerate this creature.",
			Purpose: game.Purpose{Answers: game.AnswerSurvive},
			Cost:    ManaCost("{3}"),
			Effect: func(g *game.Game, item *game.StackItem) error {
				return Regenerate{Target: item.SourceCardID}.Apply(NewContext(g, item))
			},
		}},
	})
}
