package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Martyr for the Cause — Creature — Human Soldier {1}{W}, 2/2 (War of the Spark, #5520):
//
//	"When this creature dies, proliferate. (Choose any number of permanents and/or players, then give each another counter of each kind already there.)"
//
// Blightbelly Rat's dies trigger.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "b19cbb92-bdf9-4b34-bd16-db5a594f51b7",
		Name:         "Martyr for the Cause",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			WhenThisDies("Martyr for the Cause — proliferate", Do(Proliferate{})),
		},
	})
}
