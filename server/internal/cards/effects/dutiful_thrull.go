package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Dutiful Thrull — Creature — Thrull {W}, 1/1 (Gatecrash, #5565):
//
//	"{B}: Regenerate this creature."
//
// Drowned's ability: an ordinary CR 701.19a regeneration shield.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "744fcf77-b597-48a0-92ec-f3a617c03df1",
		Name:         "Dutiful Thrull",
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
