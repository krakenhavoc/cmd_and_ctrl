package cardssh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Silver Drake — Creature — Drake {1}{W}{U}, 3/3 (Planeshift, #5615):
//
//	"Flying
//	 When this creature enters, return a white or blue creature you
//	 control to its owner's hand."
//
// Flying is a printed keyword; the enters trigger is the shared gating
// return (gating.go), chosen as it resolves, the Drake itself included.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "ec7eb7d8-92b2-4442-856e-a9790301c3b3",
		Name:            "Silver Drake",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered:       []game.TriggeredAbility{effects.GatingReturn("Silver Drake", "W", "U")},
	})
}
