package cardsw

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Wake of Vultures — Creature — Bird {3}{B}, 3/1 (Visions, #5631):
//
//	"Flying
//	 {1}{B}, Sacrifice a creature: Regenerate this creature."
//
// Flying is an engine keyword carried as printed data (CR 702.9b). The
// sacrifice is a cost, paid with any creature you control, the Vultures
// themselves included (which leaves nothing to regenerate). The shield
// works the next time the Vultures would be destroyed this turn
// (CR 701.19a).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "2948b1d5-c12c-45c4-a3e8-aa470fcee1af",
		Name:            "Wake of Vultures",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Activated: []effects.ActivatedAbility{{
			Label:   "{1}{B}, Sacrifice a creature: Regenerate this creature.",
			Purpose: game.Purpose{Answers: game.AnswerSurvive | game.AnswerSacOutlet},
			Cost:    effects.Plus(effects.ManaCost("{1}{B}"), effects.SacrificeACreature()),
			Effect: func(g *game.Game, item *game.StackItem) error {
				return effects.Regenerate{Target: item.SourceCardID}.Apply(effects.NewContext(g, item))
			},
		}},
	})
}
