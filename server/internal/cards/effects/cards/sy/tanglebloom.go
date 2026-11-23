package cardssy

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Tanglebloom — Artifact {1} (Mirrodin, #5606):
//
//	"{1}, {T}: You gain 1 life."
//
// Fountain of Youth's row at {1}: its controller gains 1 life as it
// resolves.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "87890f04-b831-4869-a7de-72789a466c71",
		Name:         "Tanglebloom",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "{1}, {T}: You gain 1 life.",
			Purpose: game.Purpose{Answers: game.AnswerValue},
			Cost:    effects.Plus(effects.ManaCost("{1}"), effects.TapCost()),
			Effect:  effects.GainOneLifeEffect,
		}},
	})
}
