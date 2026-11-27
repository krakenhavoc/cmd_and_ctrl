package cardsdi

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Dwarven Demolition Team — Creature — Dwarf {2}{R}, 1/1 (Eighth Edition,
// #5607):
//
//	"{T}: Destroy target Wall."
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "caf3c6ec-d17e-497c-8fe7-6f818ce93f96",
		Name:         "Dwarven Demolition Team",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "{T}: Destroy target Wall.",
			Purpose: game.Purpose{Answers: game.AnswerRemove},
			Cost:    effects.TapCost(),
			Targets: effects.TargetPermanent("target Wall", effects.Subtype("Wall")),
			Effect:  effects.DestroyFirstLegalTarget,
		}},
	})
}
