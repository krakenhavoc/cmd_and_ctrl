package cardslu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Metalspinner's Puzzleknot — Artifact {2} (Foundations Jumpstart):
//
//	"When this artifact enters, you draw a card and you lose 1 life.
//	 {2}{B}, Sacrifice this artifact: You draw a card and you lose 1 life."
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "c19d2f54-2c3a-4cbe-8328-061d4b51f574",
		Name:         "Metalspinner's Puzzleknot",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Draws: 1},
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("Metalspinner's Puzzleknot — you draw a card and you lose 1 life",
				effects.Do(effects.DrawCards{N: 1}, effects.YouLoseLife(1))),
		},
		Activated: []effects.ActivatedAbility{{
			Label:   "{2}{B}, Sacrifice this artifact: You draw a card and you lose 1 life.",
			Purpose: game.Purpose{Draws: 1, Answers: game.AnswerValue},
			Cost:    effects.Plus(effects.ManaCost("{2}{B}"), effects.SacrificeThis()),
			Effect:  effects.Do(effects.DrawCards{N: 1}, effects.YouLoseLife(1)),
		}},
	})
}
