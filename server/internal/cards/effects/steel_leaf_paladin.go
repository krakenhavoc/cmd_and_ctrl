package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Steel Leaf Paladin — Creature — Elf Knight {4}{G}{W}, 4/4
// (Planeshift, #5615):
//
//	"First strike
//	 When this creature enters, return a green or white creature you
//	 control to its owner's hand."
//
// First strike is a printed keyword; the enters trigger is the shared
// gating return (gating.go), chosen as it resolves, the Paladin itself
// included.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "15e0250e-a23b-499a-a8d2-24fc13084c28",
		Name:            "Steel Leaf Paladin",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"first strike"},
		Triggered:       []game.TriggeredAbility{gatingReturn("Steel Leaf Paladin", "G", "W")},
	})
}
