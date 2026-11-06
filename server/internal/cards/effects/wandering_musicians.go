package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Wandering Musicians — Creature — Human Bard Ally {3}{R/W}, 2/5 (Avatar:
// The Last Airbender, #4459):
//
//	"Whenever this creature attacks, creatures you control get +1/+0
//	 until end of turn."
//
// The pump is Master of Barbs': the creatures you control as the trigger
// resolves, locked in then (CR 611.2c).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "7166bf81-05c2-457d-a0bb-a175175b8e79",
		Name:         "Wandering Musicians",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			WheneverThisAttacks("Wandering Musicians — creatures you control get +1/+0 until end of turn",
				Do(BoostUntilEOT{Match: And(Creature(), YouControl()), Power: 1, Label: "Wandering Musicians — +1/+0"})),
		},
	})
}
