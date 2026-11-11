package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Aetherborn Marauder — Creature — Aetherborn Rogue {3}{B}, 2/2 (#5368):
//
//	"Flying, lifelink
//	 When this creature enters, move any number of +1/+1 counters from other
//	 permanents you control onto this creature."
//
// Moving counters (CR 122.5) takes them off one permanent and puts them on
// another. "Any number" is asked of the controller as the trigger resolves
// (CR 608.2c), one source permanent at a time in battlefield order, zero
// being allowed; the removal and the placement go through the same counter
// primitive, so a doubler sees the placement. Only the other permanents the
// controller controls with +1/+1 counters are asked. Nothing moves if the
// Marauder has left the battlefield.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "ff6427f6-d2b1-4e33-984d-c06bd15d7c2c",
		Name:            "Aetherborn Marauder",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying", "lifelink"},
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("Aetherborn Marauder — move any number of +1/+1 counters from other permanents you control onto this creature",
				cmdr2B23MarauderMove),
		},
	})
}
