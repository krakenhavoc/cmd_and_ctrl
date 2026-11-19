package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Whirlermaker — Artifact {3} (Foundations Jumpstart):
//
//	"{4}, {T}: Create a 1/1 colorless Thopter artifact creature token with
//	 flying."
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "bdf28770-0962-4317-822f-c187bc6706d1",
		Name:         "Whirlermaker",
		Completeness: CompletenessFull,
		Activated: []ActivatedAbility{{
			Label:   "{4}, {T}: Create a 1/1 colorless Thopter artifact creature token with flying.",
			Purpose: game.Purpose{Tokens: 1, CreatureTokens: 1, Answers: game.AnswerValue},
			Cost:    Plus(ManaCost("{4}"), TapCost()),
			Effect:  Do(CreateToken{Template: TokenCard("1/1 colorless Thopter artifact with flying"), N: 1}),
		}},
	})
}
