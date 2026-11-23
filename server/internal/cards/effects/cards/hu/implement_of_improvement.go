package cardshu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Implement of Improvement — Artifact {1} (Aether Revolt,
// #5536):
//
//	"{W}, Sacrifice this artifact: You gain 2 life.
//	 When this artifact is put into a graveyard from the battlefield,
//	 draw a card."
//
// A Food-shaped life row at instant speed.
// The draw is Chromatic Star's leaves-for-the-graveyard trigger: it
// fires however the Implement got there, sacrificed to its own ability
// or destroyed.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "a71b2ffb-6a62-4239-987f-ec82a358927f",
		Name:         "Implement of Improvement",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "{W}, Sacrifice this artifact: You gain 2 life.",
			Purpose: game.Purpose{Answers: game.AnswerValue},
			Cost:    effects.Plus(effects.ManaCost("{W}"), effects.SacrificeThis()),
			Effect:  effects.Do(effects.GainLife{Amount: 2}),
		}},
		Triggered: []game.TriggeredAbility{
			effects.WhenThisDies("Implement of Improvement — draw a card", effects.Do(effects.DrawCards{N: 1})),
		},
	})
}
