package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Fearsome Goblin Pair — Creature — Goblin Soldier {2}{B/R}, 1/1 (The Hobbit):
//
//	"When this creature dies, amass Goblins 4."
//
// The dies trigger is the ordinary amass: find the controller's Army
// or make a 0/0 Goblin Army, then four +1/+1 counters (CR 701.47a).
// The trigger's controller is the creature's last controller.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "edbd3d4e-78a2-4c03-936d-dcb24cf7c87c",
		Name:         "Fearsome Goblin Pair",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			WhenThisDies("Fearsome Goblin Pair — amass Goblins 4", Do(Amass{Subtype: "Goblin", N: 4})),
		},
	})
}
