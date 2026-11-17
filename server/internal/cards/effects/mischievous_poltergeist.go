package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Mischievous Poltergeist — Creature — Spirit {2}{B}, 1/1 (Weatherlight, #5629):
//
//	"Flying
//	 Pay 1 life: Regenerate this creature."
//
// Flying is the printed keyword. The ability is River Boa's
// regeneration shield (CR 701.19a) for 1 life, which a player can pay
// only with at least 1 life (CR 119.4).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "c12b5b9a-a46e-4635-8603-62c2ca5ee58e",
		Name:            "Mischievous Poltergeist",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Activated: []ActivatedAbility{{
			Label:   "Pay 1 life: Regenerate this creature.",
			Purpose: game.Purpose{Answers: game.AnswerSurvive},
			Cost:    PayLife(1),
			Effect: func(g *game.Game, item *game.StackItem) error {
				return Regenerate{Target: item.SourceCardID}.Apply(NewContext(g, item))
			},
		}},
	})
}
