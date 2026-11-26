package cardsco

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Cogworker's Puzzleknot — Artifact {2} (Kaladesh, #5539):
//
//	"When this artifact enters, create a 1/1 colorless Servo artifact creature
//	 token.
//	 {1}{W}, Sacrifice this artifact: Create a 1/1 colorless Servo artifact
//	 creature token."
//
// Inspired Inventor's Servo on the entry trigger, and the same token as the
// effect of an ability whose cost sacrifices the Puzzleknot itself (CR 602.2b),
// so the Servo is made even though the artifact is gone.
//
// No simplification.
func init() {
	servo := effects.Do(effects.CreateToken{Template: effects.TokenCard("1/1 colorless Servo artifact"), N: 1})
	effects.Register(effects.Spec{
		OracleID:     "2065e9d0-c6e9-4270-898e-9822b532d467",
		Name:         "Cogworker's Puzzleknot",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Tokens: 1, CreatureTokens: 1},
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("Cogworker's Puzzleknot — create a 1/1 colorless Servo artifact creature token", servo),
		},
		Activated: []effects.ActivatedAbility{{
			Label:   "{1}{W}, Sacrifice this artifact: Create a 1/1 colorless Servo artifact creature token.",
			Purpose: game.Purpose{Answers: game.AnswerMakesBlocker, Tokens: 1, CreatureTokens: 1},
			Cost:    effects.Plus(effects.ManaCost("{1}{W}"), effects.SacrificeThis()),
			Effect:  servo,
		}},
	})
}
