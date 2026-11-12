package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Ghost Ship — Creature — Spirit {2}{U}{U}, 2/4 (The Dark, #5638):
//
//	"Flying
//	 {U}{U}{U}: Regenerate this creature."
//
// A printed keyword and Albino Troll's regeneration row with a
// {U}{U}{U} cost.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "09de1136-8a8f-4539-92a2-611dcdd0790b",
		Name:            "Ghost Ship",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Activated: []ActivatedAbility{{
			Label:   "{U}{U}{U}: Regenerate this creature.",
			Purpose: game.Purpose{Answers: game.AnswerSurvive},
			Cost:    ManaCost("{U}{U}{U}"),
			Effect: func(g *game.Game, item *game.StackItem) error {
				return Regenerate{Target: item.SourceCardID}.Apply(NewContext(g, item))
			},
		}},
	})
}
