package cardsr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Ranger en-Vec — Creature — Human Soldier Archer Ranger {1}{G}{W}, 2/2 (Tempest,
// S82 OLD b752, #5628):
//
//	"First strike
//	 {G}: Regenerate this creature."
//
// First strike is an engine keyword carried as printed data (CR 702.7b). The
// ability is Albino Troll's: Regenerate on the ability's own source (CR 701.19).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "04bb6731-5662-48d5-9e48-99d378c4adaf",
		Name:            "Ranger en-Vec",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"first strike"},
		Activated: []effects.ActivatedAbility{{
			Label:   "{G}: Regenerate this creature.",
			Purpose: game.Purpose{Answers: game.AnswerSurvive},
			Cost:    effects.ManaCost("{G}"),
			Effect: func(g *game.Game, item *game.StackItem) error {
				return effects.Regenerate{Target: item.SourceCardID}.Apply(effects.NewContext(g, item))
			},
		}},
	})
}
