package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	Register(Spec{
		OracleID:     "3b02dfca-bcaa-4e22-b143-401ed250d016",
		Name:         "Titan Forge",
		Completeness: CompletenessFull,
		Activated: []ActivatedAbility{
			{
				Label:   "{3}, {T}: Put a charge counter on this artifact.",
				Purpose: game.Purpose{Answers: game.AnswerValue},
				Cost:    Plus(ManaCost("{3}"), TapCost()),
				Effect:  b33PutChargeCounterOnSelf,
			},
			{
				Label:   "{T}, Remove three charge counters from this artifact: Create a 9/9 colorless Golem artifact creature token.",
				Purpose: game.Purpose{Tokens: 1, CreatureTokens: 1, Answers: game.AnswerMakesBlocker},
				Cost:    Plus(TapCost(), RemoveCountersFromThis(game.CounterCharge, 3)),
				Effect:  Do(CreateToken{Template: TokenCard("9/9 colorless Golem artifact creature"), N: 1}),
			},
		},
	})
}
