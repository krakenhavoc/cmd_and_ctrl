package cardsri

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Rusted Slasher — Artifact Creature — Phyrexian Horror {4}, 4/1 (Mirrodin
// Besieged, S82 OLD b465, #5576):
//
//	"Sacrifice an artifact: Regenerate this creature."
//
// Atog's cost (one artifact the controller picks at announce, the Slasher
// itself included) with Regenerate on the ability's own source (CR 701.19).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "098cc111-493b-470b-929c-67deb981240c",
		Name:         "Rusted Slasher",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "Sacrifice an artifact: Regenerate this creature.",
			Purpose: game.Purpose{Answers: game.AnswerSurvive},
			Cost:    effects.SacrificeN(1, "an artifact", effects.Artifact()),
			Effect: func(g *game.Game, item *game.StackItem) error {
				return effects.Regenerate{Target: item.SourceCardID}.Apply(effects.NewContext(g, item))
			},
		}},
	})
}
