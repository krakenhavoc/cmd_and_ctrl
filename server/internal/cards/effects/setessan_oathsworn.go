package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Setessan Oathsworn — Creature — Satyr Warrior {1}{G}{G}, 1/1 (Born of the Gods,
// #5559):
//
//	"Heroic — Whenever you cast a spell that targets this creature,
//	 put two +1/+1 counters on this creature."
//
// The heroic seam (heroic) with plusOneCountersOnThis(2).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "78d5f882-5d17-4228-b9b5-d460ab2bef08",
		Name:         "Setessan Oathsworn",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			heroic("Setessan Oathsworn — two +1/+1 counters on this creature (heroic)", plusOneCountersOnThis(2)),
		},
	})
}
