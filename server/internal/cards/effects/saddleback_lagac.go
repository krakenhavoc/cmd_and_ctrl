package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Saddleback Lagac — Creature — Lizard {3}{G}, 3/1 (Oath of the Gatewatch, #5545):
//
//	"When this creature enters, support 2. (Put a +1/+1 counter on each
//	 of up to two other target creatures.)"
//
// Support from a permanent targets OTHER creatures (CR 701.41a): the
// shared WhenThisEntersSupport.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "3d720d55-3df8-40fd-bc93-ad99c1105249",
		Name:         "Saddleback Lagac",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			WhenThisEntersSupport("Saddleback Lagac", 2, true),
		},
	})
}
