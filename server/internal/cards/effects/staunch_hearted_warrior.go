package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Staunch-Hearted Warrior — Creature — Human Warrior {3}{G}, 2/2 (Theros, #5561):
//
//	"Heroic — Whenever you cast a spell that targets this creature, put two +1/+1 counters on this creature."
//
// The shared heroic trigger (CR 207.2c). The counters land only on the same object.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "aa6244fb-e1f1-4b53-b530-a8c9d0712a5b",
		Name:         "Staunch-Hearted Warrior",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			heroic("Staunch-Hearted Warrior — two +1/+1 counters on it (heroic)", thsPlusOneCountersOnThisIfSameObject(2)),
		},
	})
}
