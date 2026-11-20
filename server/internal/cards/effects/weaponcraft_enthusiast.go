package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Weaponcraft Enthusiast — Creature — Aetherborn Artificer {2}{B}, 0/1 (Kaladesh, #5539):
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
		OracleID:     "7c27136a-bf9d-4f42-9c37-abe347fad1bb",
		Name:         "Weaponcraft Enthusiast",
		Completeness: CompletenessFull,
		Triggered:    []game.TriggeredAbility{Fabricate("Weaponcraft Enthusiast", 2)},
	})
}
