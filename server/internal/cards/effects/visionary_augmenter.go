package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Visionary Augmenter — Creature — Dwarf Artificer {2}{W}{W}, 2/1 (Kaladesh, #5539):
//
//	"Fabricate 2 (When this creature enters, put two +1/+1 counters on
//	 it or create two 1/1 colorless Servo artifact creature tokens.)"
//
// Fabricate is the shared keyword constructor (fabricate_keyword.go,
// CR 702.123a): the enters trigger asks counters or Servos as it
// resolves, and makes the Servos without asking once this creature has
// left the battlefield.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "1782bfd7-7d8f-4ed2-9d45-ba9adb41e996",
		Name:         "Visionary Augmenter",
		Completeness: CompletenessFull,
		Triggered:    []game.TriggeredAbility{Fabricate("Visionary Augmenter", 2)},
	})
}
