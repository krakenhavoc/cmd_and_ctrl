package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Mechanized Ninja Cavalry — {1}{R/W} Artifact Creature — Robot Ninja 1/1
// (Teenage Mutant Ninja Turtles, #4457):
//
//	"When this creature enters, create a 1/1 colorless Robot artifact
//	 creature token."
//
// The set's 1/1 Robot (Ravenous Robots' row).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "2db8dda3-8136-4cf2-ba2d-ecc344a36725",
		Name:         "Mechanized Ninja Cavalry",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("Mechanized Ninja Cavalry — create a 1/1 colorless Robot artifact creature token",
				Do(CreateToken{Template: TokenCard("1/1 colorless Robot artifact"), N: 1})),
		},
	})
}
