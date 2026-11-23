package cardsg

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

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
	effects.Register(effects.Spec{
		OracleID:        "09de1136-8a8f-4539-92a2-611dcdd0790b",
		Name:            "Ghost Ship",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Activated: []effects.ActivatedAbility{{
			Label:   "{U}{U}{U}: Regenerate this creature.",
			Purpose: game.Purpose{Answers: game.AnswerSurvive},
			Cost:    effects.ManaCost("{U}{U}{U}"),
			Effect: func(g *game.Game, item *game.StackItem) error {
				return effects.Regenerate{Target: item.SourceCardID}.Apply(effects.NewContext(g, item))
			},
		}},
	})
}
