package cardsv

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Viashino Sandswimmer — Creature — Lizard {2}{R}{R} 3/2 (Urza's Saga, #5624):
//
//	"{R}: Flip a coin. If you win the flip, return this creature to its
//	 owner's hand. If you lose the flip, sacrifice this creature."
//
// The coin is called by the controller (CR 705.2); the result is applied to
// the same object if it is still on the battlefield.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "1d82f01a-29be-4756-8745-e2c04bf77c61",
		Name:         "Viashino Sandswimmer",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "{R}: Flip a coin. If you win the flip, return this creature to its owner's hand. If you lose the flip, sacrifice this creature.",
			Purpose: game.Purpose{Answers: game.AnswerValue},
			Cost:    effects.ManaCost("{R}"),
			Effect:  effects.S82OldB726FlipToBounceOrSacrifice,
		}},
	})
}
