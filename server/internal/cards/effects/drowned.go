package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Drowned — Creature — Zombie {1}{U}, 1/1 (The Dark, #5638):
//
//	"{B}: Regenerate this creature."
//
// Albino Troll's regeneration row with a {B} cost: Regenerate on the
// ability's own source, a shield that lasts until the turn ends.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "3da87e58-6a4b-48c3-873c-bc654277d4df",
		Name:         "Drowned",
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
