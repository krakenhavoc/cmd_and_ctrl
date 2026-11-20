package cardssy

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Tenacious Drone — Artifact Creature — Borg {1}{B}, 2/1 (Star Trek,
// #4973):
//
//	"When this creature dies, create a 1/1 colorless Borg artifact
//	 creature token."
//
// A leaves-the-battlefield trigger that looks back in time (CR 603.10a):
// it fires from the graveyard move and makes the token for the Drone's
// controller as it last existed on the battlefield. The token is the
// printed ttrk Borg, a plain 1/1 colourless artifact creature.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "af6351a9-ac01-4fed-9ea9-a352d21a8312",
		Name:         "Tenacious Drone",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.WhenThisDies("Tenacious Drone — create a 1/1 colorless Borg artifact creature token",
				effects.Do(effects.CreateToken{Template: effects.TokenCard("1/1 colorless Borg artifact"), N: 1})),
		},
	})
}
