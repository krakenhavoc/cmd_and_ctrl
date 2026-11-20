package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Accomplished Automaton — Artifact Creature — Construct {7}, 5/7 (Kaladesh, #5539):
//
//	"Fabricate 1 (When this creature enters, put a +1/+1 counter on it
//	 or create a 1/1 colorless Servo artifact creature token.)"
//
// Fabricate is the shared keyword constructor (fabricate_keyword.go,
// CR 702.123a): the enters trigger asks counters or Servos as it
// resolves, and makes the Servos without asking once this creature has
// left the battlefield.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "f5af0d1b-531e-4053-848a-5e0fa1022f9a",
		Name:         "Accomplished Automaton",
		Completeness: CompletenessFull,
		Triggered:    []game.TriggeredAbility{Fabricate("Accomplished Automaton", 1)},
	})
}
