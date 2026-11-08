package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Fortress Kin-Guard — Creature — Dog Soldier {1}{W}, 1/2 (Tarkir:
// Dragonstorm, #4463):
//
//	"When this creature enters, it endures 1. (Put a +1/+1 counter on it
//	 or create a 1/1 white Spirit creature token.)"
//
// Endure is endure_n.go (CR 701.63a), Sandskitter Outrider's entry
// trigger.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "fa3ed0f6-e44e-433d-80d3-824de5e6135a",
		Name:         "Fortress Kin-Guard",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("Fortress Kin-Guard — it endures 1", ThisEndures("Fortress Kin-Guard", 1)),
		},
	})
}
