package cardsfl

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Fountain of Youth — Artifact {0} (The Dark, #5638):
//
//	"{2}, {T}: You gain 1 life."
//
// Overgrown Arch's life row with a {2} component. The gain uses the
// stack and can be responded to.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "b906923c-9997-4da5-a05e-53eb4d2dff32",
		Name:         "Fountain of Youth",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "{2}, {T}: You gain 1 life.",
			Purpose: game.Purpose{Answers: game.AnswerValue},
			Cost:    effects.Plus(effects.ManaCost("{2}"), effects.TapCost()),
			Effect:  effects.GainOneLifeEffect,
		}},
	})
}
