package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Servo Schematic — Artifact {2}:
//
//	"When this artifact enters or is put into a graveyard from the
//	 battlefield, create a 1/1 colorless Servo artifact creature token."
//
// One printed triggered ability with two conditions (WhenThisEntersOrDies,
// as Mogg War Marshal): the Schematic entering, or going from the
// battlefield to a graveyard. A Schematic exiled or bounced makes no
// second Servo.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "6b8c6104-c537-4025-83e2-e89c8668f3ba",
		Name:         "Servo Schematic",
		Purpose:      game.Purpose{Tokens: 1},
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			WhenThisEntersOrDies("Servo Schematic — create a 1/1 Servo",
				Do(CreateToken{Template: TokenCard("1/1 colorless Servo artifact"), N: 1})),
		},
	})
}
