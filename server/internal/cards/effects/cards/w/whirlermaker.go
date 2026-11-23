package cardsw

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Whirlermaker — Artifact {3} (Foundations Jumpstart):
//
//	"{4}, {T}: Create a 1/1 colorless Thopter artifact creature token with
//	 flying."
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "bdf28770-0962-4317-822f-c187bc6706d1",
		Name:         "Whirlermaker",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "{4}, {T}: Create a 1/1 colorless Thopter artifact creature token with flying.",
			Purpose: game.Purpose{Tokens: 1, CreatureTokens: 1, Answers: game.AnswerValue},
			Cost:    effects.Plus(effects.ManaCost("{4}"), effects.TapCost()),
			Effect:  effects.Do(effects.CreateToken{Template: effects.TokenCard("1/1 colorless Thopter artifact with flying"), N: 1}),
		}},
	})
}
