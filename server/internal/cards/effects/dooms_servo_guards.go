package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Doom's Servo-Guards — Artifact Creature — Robot Villain {1}{B}, 2/2
// (Marvel Super Heroes Commander, #4455): "When this creature enters, you
// gain 2 life and mill two cards." Life first, then the mill (CR 701.17a).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "296e79f9-57cb-4af7-9e18-48d2d655b537",
		Name:         "Doom's Servo-Guards",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("Doom's Servo-Guards — gain 2 life and mill two cards",
				Do(GainLife{Amount: 2}, MillCards{N: 2})),
		},
	})
}
