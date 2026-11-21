package cardsp

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Potion of Healing — Artifact {1}{W} (AFR):
//
//	"When this artifact enters, draw a card.
//	 {W}, {T}, Sacrifice this artifact: You gain 3 life."
func init() {
	effects.Register(effects.Spec{
		OracleID:     "11fb3fee-ae84-4bf3-b834-63d3b8987df3",
		Name:         "Potion of Healing",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("Potion of Healing — draw a card", effects.Do(effects.DrawCards{N: 1})),
		},
		Activated: []effects.ActivatedAbility{{
			Label:   "{W}, {T}, Sacrifice this artifact: You gain 3 life.",
			Purpose: game.Purpose{LifeGain: 3, Answers: game.AnswerValue},
			Cost:    game.AbilityCost{Mana: "{W}", Tap: true, SacrificeSelf: true},
			Effect:  effects.Do(effects.GainLife{Amount: 3}),
		}},
	})
}
