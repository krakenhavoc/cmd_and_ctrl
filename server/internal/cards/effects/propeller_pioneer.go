package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Propeller Pioneer — Creature — Human Artificer {3}{W}, 2/1 (Kaladesh, #5539):
//
//	"Flying
//	 Fabricate 1 (When this creature enters, put a +1/+1 counter on it
//	 or create a 1/1 colorless Servo artifact creature token.)"
//
// The printed keyword is read by the engine. Fabricate is the shared keyword constructor (fabricate_keyword.go,
// CR 702.123a): the enters trigger asks counters or Servos as it
// resolves, and makes the Servos without asking once this creature has
// left the battlefield.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "d17a9750-a89f-4ad1-9c15-7ccffc554ee5",
		Name:            "Propeller Pioneer",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered:       []game.TriggeredAbility{Fabricate("Propeller Pioneer", 1)},
	})
}
