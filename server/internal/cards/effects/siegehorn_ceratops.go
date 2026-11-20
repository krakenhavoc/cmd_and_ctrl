package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Siegehorn Ceratops — Creature — Dinosaur {G}{W}, 2/2 (Rivals of Ixalan,
// #5531):
//
//	"Enrage — Whenever this creature is dealt damage, put two +1/+1
//	 counters on it. (It must survive the damage to get the counters.)"
//
// The shared enrage trigger with plusOneCountersOnThis: a Ceratops that
// died to the damage is gone by resolution, and the counters go nowhere,
// as the reminder says.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "f7730e2e-aa62-40e7-836d-e6ddb6893816",
		Name:         "Siegehorn Ceratops",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			cmdr2B08EnrageTrigger("Siegehorn Ceratops — Enrage: put two +1/+1 counters on it", plusOneCountersOnThis(2)),
		},
	})
}
