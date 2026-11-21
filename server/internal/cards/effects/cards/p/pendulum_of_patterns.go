package cardsp

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Pendulum of Patterns — Artifact {2} (Aether Revolt,
// #5536):
//
//	"When this artifact enters, you gain 3 life.
//	 {5}, {T}, Sacrifice this artifact: Draw a card."
//
// An enters trigger on the stack and a tap-and-sacrifice draw row.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "992075d5-f413-4896-b75e-7bb2d589c50e",
		Name:         "Pendulum of Patterns",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("Pendulum of Patterns — you gain 3 life", effects.Do(effects.GainLife{Amount: 3})),
		},
		Activated: []effects.ActivatedAbility{{
			Label:   "{5}, {T}, Sacrifice this artifact: Draw a card.",
			Purpose: game.Purpose{Answers: game.AnswerValue, Draws: 1},
			Cost:    effects.Plus(effects.ManaCost("{5}"), effects.TapCost(), effects.SacrificeThis()),
			Effect:  effects.Do(effects.DrawCards{N: 1}),
		}},
	})
}
