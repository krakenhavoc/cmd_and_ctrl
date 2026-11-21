package cardsfl

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Fleetfoot Panther — Creature — Cat {1}{G}{W}, 3/4 (Planeshift, #5615):
//
//	"Flash
//	 When this creature enters, return a green or white creature you
//	 control to its owner's hand."
//
// Flash is a printed keyword, so the Panther can save a green or white
// creature from removal at instant speed. The enters trigger is the
// shared gating return (gating.go), chosen as it resolves, the Panther
// itself included.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "d99e8f94-4a8a-42a1-b5d0-2b08cacb8676",
		Name:            "Fleetfoot Panther",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flash"},
		Triggered:       []game.TriggeredAbility{effects.GatingReturn("Fleetfoot Panther", "G", "W")},
	})
}
