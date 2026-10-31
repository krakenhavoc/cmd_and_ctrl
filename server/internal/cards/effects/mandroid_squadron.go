package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Mandroid Squadron — Artifact Creature — Human Soldier {1}{W}, 0/4
// (Marvel Super Heroes Commander, #4455): "When this creature enters,
// you gain 2 life." The ETB gain-life shape of Doom's Servo-Guards.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "95ab1529-0de0-416a-82b4-c5ed9276a9b4",
		Name:         "Mandroid Squadron",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("Mandroid Squadron — you gain 2 life", Do(GainLife{Amount: 2})),
		},
	})
}
