package cardsdi

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Doom's Servo-Guards — Artifact Creature — Robot Villain {1}{B}, 2/2
// (Marvel Super Heroes Commander, #4455): "When this creature enters, you
// gain 2 life and mill two cards." Life first, then the mill (CR 701.17a).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "296e79f9-57cb-4af7-9e18-48d2d655b537",
		Name:         "Doom's Servo-Guards",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("Doom's Servo-Guards — gain 2 life and mill two cards",
				effects.Do(effects.GainLife{Amount: 2}, effects.MillCards{N: 2})),
		},
	})
}
