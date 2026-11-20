package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Ambitious Aetherborn — Creature — Aetherborn Artificer {4}{B}, 4/3 (Kaladesh, #5539):
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
		OracleID:     "31f54f88-85a6-43f4-a6de-fb6ef5f0c38d",
		Name:         "Ambitious Aetherborn",
		Completeness: CompletenessFull,
		Triggered:    []game.TriggeredAbility{Fabricate("Ambitious Aetherborn", 1)},
	})
}
