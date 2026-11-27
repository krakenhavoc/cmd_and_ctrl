package cardslu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Marsh Lurker — Creature — Beast {3}{B}, 3/2 (Tempest, S82 OLD b753, #5628):
//
//	"Sacrifice a Swamp: This creature gains fear until end of turn."
//
// Rootwater Alligator's cost with a grant on the Lurker itself.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "4b35b928-1088-457d-bd0d-3b14dc4ab230",
		Name:         "Marsh Lurker",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "Sacrifice a Swamp: This creature gains fear until end of turn.",
			Purpose: game.Purpose{Answers: game.AnswerSacOutlet, Pump: &game.Pump{Keywords: []string{"fear"}}},
			Cost:    effects.SacrificeN(1, "a Swamp", effects.Subtype("Swamp")),
			Effect:  effects.ThisGainsKeywordUntilEndOfTurn("fear", "Marsh Lurker — fear until end of turn"),
		}},
	})
}
