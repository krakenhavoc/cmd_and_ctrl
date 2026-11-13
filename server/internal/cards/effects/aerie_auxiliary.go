package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Aerie Auxiliary — Creature — Bird Soldier {3}{W}, 3/3 (Modern Horizons
// 3, #5467):
//
//	"Flying
//	 When this creature enters, support 2. (Put a +1/+1 counter on each
//	 of up to two other target creatures.)"
//
// Generous Patron's support (CR 701.41a): on a creature it is "other
// target creatures", so the Auxiliary can't pick itself, and a target
// gone by resolution takes nothing (CR 608.2b). Flying is the engine's
// keyword.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "0c97aea7-c997-45b4-a044-84fb5c07a021",
		Name:            "Aerie Auxiliary",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			WhenThisEntersSupport("Aerie Auxiliary", 2, true),
		},
	})
}
