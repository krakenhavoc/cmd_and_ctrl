package cardsg

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Gluttonous Cyclops — Creature — Cyclops {5}{R}, 5/4 (Journey into Nyx, S82
// OLD, #5558):
//
//	"{5}{R}{R}: Monstrosity 3. (If this creature isn't monstrous, put three
//	 +1/+1 counters on it and it becomes monstrous.)"
//
// The monstrosity constructor alone (Fleecemane Lion's activated half).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "35101c6c-0bcc-4b1f-9cb5-a82e4d2f8220",
		Name:         "Gluttonous Cyclops",
		Completeness: effects.CompletenessFull,
		Activated:    []effects.ActivatedAbility{effects.Monstrosity(effects.ManaCost("{5}{R}{R}"), 3)},
	})
}
