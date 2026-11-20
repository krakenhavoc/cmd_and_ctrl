package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Centaur Battlemaster — Creature — Centaur Warrior {3}{G}{G}, 3/3 (Theros, #5561):
//
//	"Heroic — Whenever you cast a spell that targets this creature, put three +1/+1 counters on this creature."
//
// The shared heroic trigger (CR 207.2c). The counters land only on the same object.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "4b854749-f3a5-4199-bbfa-83090b8ccc25",
		Name:         "Centaur Battlemaster",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			heroic("Centaur Battlemaster — three +1/+1 counters on it (heroic)", thsPlusOneCountersOnThisIfSameObject(3)),
		},
	})
}
