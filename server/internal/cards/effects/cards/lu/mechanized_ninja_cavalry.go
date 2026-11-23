package cardslu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

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
	effects.Register(effects.Spec{
		OracleID:     "2db8dda3-8136-4cf2-ba2d-ecc344a36725",
		Name:         "Mechanized Ninja Cavalry",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("Mechanized Ninja Cavalry — create a 1/1 colorless Robot artifact creature token",
				effects.Do(effects.CreateToken{Template: effects.TokenCard("1/1 colorless Robot artifact"), N: 1})),
		},
	})
}
