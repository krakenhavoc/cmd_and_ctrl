package cardsw

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Witching Well — Artifact {U} (Foundations Jumpstart):
//
//	"When this artifact enters, scry 2.
//	 {3}{U}, Sacrifice this artifact: Draw two cards."
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "baadfc17-42fd-4771-aabe-027ef7b3bf69",
		Name:         "Witching Well",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Scry: 2},
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("Witching Well — scry 2", effects.Do(effects.Scry{N: 2})),
		},
		Activated: []effects.ActivatedAbility{{
			Label:   "{3}{U}, Sacrifice this artifact: Draw two cards.",
			Purpose: game.Purpose{Draws: 2, Answers: game.AnswerValue},
			Cost:    effects.Plus(effects.ManaCost("{3}{U}"), effects.SacrificeThis()),
			Effect:  effects.Do(effects.DrawCards{N: 2}),
		}},
	})
}
