package cardsth

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Titan Forge — Artifact {3} (Foundations Commander):
//
//	"{3}, {T}: Put a charge counter on this artifact.
//	 {T}, Remove three charge counters from this artifact: Create a 9/9
//	 colorless Golem artifact creature token."
//
// Magistrate's Scepter's two abilities with a 9/9 Golem for the payoff.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "3b02dfca-bcaa-4e22-b143-401ed250d016",
		Name:         "Titan Forge",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{
			{
				Label:   "{3}, {T}: Put a charge counter on this artifact.",
				Purpose: game.Purpose{Answers: game.AnswerValue},
				Cost:    effects.Plus(effects.ManaCost("{3}"), effects.TapCost()),
				Effect:  effects.B33PutChargeCounterOnSelf,
			},
			{
				Label:   "{T}, Remove three charge counters from this artifact: Create a 9/9 colorless Golem artifact creature token.",
				Purpose: game.Purpose{Tokens: 1, CreatureTokens: 1, Answers: game.AnswerMakesBlocker},
				Cost:    effects.Plus(effects.TapCost(), effects.RemoveCountersFromThis(game.CounterCharge, 3)),
				Effect:  effects.Do(effects.CreateToken{Template: effects.TokenCard("9/9 colorless Golem artifact creature"), N: 1}),
			},
		},
	})
}
