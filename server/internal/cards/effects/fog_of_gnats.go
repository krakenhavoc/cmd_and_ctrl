package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Fog of Gnats — Creature — Insect {B}{B}, 1/1 (Urza's Legacy, #5623):
//
//	"Flying
//	 {B}: Regenerate this creature."
//
// Yavimaya Gnats with a {B} cost: flying is a printed keyword (CR 702.9b),
// and the shield replaces its next destruction this turn (CR 701.19a).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "41f58e06-5ce0-4b2d-a280-5bb3b6c56dfe",
		Name:            "Fog of Gnats",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying"},
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
