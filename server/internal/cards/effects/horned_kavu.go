package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Horned Kavu — Creature — Kavu {R}{G}, 3/4 (Planeshift, #5615):
//
//	"When this creature enters, return a red or green creature you
//	 control to its owner's hand."
//
// The shared gating return (gating.go), chosen as it resolves, the Kavu
// itself included.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "14817050-9f2d-453f-b36b-e6429fa29165",
		Name:         "Horned Kavu",
		Completeness: CompletenessFull,
		Triggered:    []game.TriggeredAbility{gatingReturn("Horned Kavu", "R", "G")},
	})
}
