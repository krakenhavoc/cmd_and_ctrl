package cardsh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Hidden Stockpile — Enchantment {W}{B}:
//
//	"Revolt — At the beginning of your end step, if a permanent left
//	 the battlefield under your control this turn, create a 1/1
//	 colorless Servo artifact creature token.
//	 {1}, Sacrifice a creature: Scry 1."
//
// Revolt is an intervening if (revolt.go). The sacrifice is a cost, so
// the sacrificed creature itself counts toward revolt at the end step.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "f5ded323-75c8-471e-a516-238b9d6b06d3",
		Name:         "Hidden Stockpile",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.AtYourEndStepIfRevolt("Hidden Stockpile — create a 1/1 colorless Servo artifact creature token",
				effects.Do(effects.CreateToken{Template: effects.TokenCard("1/1 colorless Servo artifact"), N: 1})),
		},
		Activated: []effects.ActivatedAbility{{
			Label:   "{1}, Sacrifice a creature: Scry 1.",
			Purpose: game.Purpose{Answers: game.AnswerSacOutlet},
			Cost:    effects.Plus(effects.ManaCost("{1}"), effects.SacrificeACreature()),
			Effect: func(g *game.Game, item *game.StackItem) error {
				return effects.Scry{Player: item.Controller, N: 1}.Apply(effects.NewContext(g, item))
			},
		}},
	})
}
