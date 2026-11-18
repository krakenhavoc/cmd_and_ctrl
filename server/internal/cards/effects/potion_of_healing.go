package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Potion of Healing — Artifact {1}{W} (AFR):
//
//	"When this artifact enters, draw a card.
//	 {W}, {T}, Sacrifice this artifact: You gain 3 life."
func init() {
	Register(Spec{
		OracleID:     "11fb3fee-ae84-4bf3-b834-63d3b8987df3",
		Name:         "Potion of Healing",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("Potion of Healing — draw a card", Do(DrawCards{N: 1})),
		},
		Activated: []ActivatedAbility{{
			Label:   "{W}, {T}, Sacrifice this artifact: You gain 3 life.",
			Purpose: game.Purpose{LifeGain: 3, Answers: game.AnswerValue},
			Cost:    game.AbilityCost{Mana: "{W}", Tap: true, SacrificeSelf: true},
			Effect:  Do(GainLife{Amount: 3}),
		}},
	})
}
