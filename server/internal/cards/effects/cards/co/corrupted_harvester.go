package cardsco

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Corrupted Harvester — Creature — Phyrexian Horror {4}{B}{B}, 6/3 (Scars of Mirrodin, #5577):
//
//	"{B}, Sacrifice a creature: Regenerate this creature."
//
// Wake of Vultures' ability at {B}. The Harvester may sacrifice itself to
// pay the cost, which regenerates nothing.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "f105ccdd-65bc-4880-bbfe-f57fc0e060ea",
		Name:         "Corrupted Harvester",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "{B}, Sacrifice a creature: Regenerate this creature.",
			Purpose: game.Purpose{Answers: game.AnswerSurvive | game.AnswerSacOutlet},
			Cost:    effects.Plus(effects.ManaCost("{B}"), effects.SacrificeACreature()),
			Effect: func(g *game.Game, item *game.StackItem) error {
				return effects.Regenerate{Target: item.SourceCardID}.Apply(effects.NewContext(g, item))
			},
		}},
	})
}
