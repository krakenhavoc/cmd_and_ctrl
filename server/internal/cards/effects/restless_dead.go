package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Restless Dead — Creature — Skeleton {1}{B}, 1/1 (Mirage, #5633):
//
//	"{B}: Regenerate this creature."
//
// Albino Troll's row at {B}: Regenerate on the ability's own source, a
// shield that replaces the next destruction this turn (CR 701.19a).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "5315ba59-ca80-498e-b21a-ed87e6a4c955",
		Name:         "Restless Dead",
		Completeness: CompletenessFull,
		Activated: []ActivatedAbility{{
			Label:   "{B}: Regenerate this creature.",
			Purpose: game.Purpose{Answers: game.AnswerSurvive},
			Cost:    ManaCost("{B}"),
			Effect: func(g *game.Game, item *game.StackItem) error {
				return Regenerate{Target: item.SourceCardID}.Apply(NewContext(g, item))
			},
		}},
	})
}
