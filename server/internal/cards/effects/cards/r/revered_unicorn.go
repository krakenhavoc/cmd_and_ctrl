package cardsr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Revered Unicorn — {1}{W} Creature — Unicorn, 2/3 (Weatherlight, S82 OLD batch
// 772, #5629):
//
//	"Cumulative upkeep {1}
//	 When this creature leaves the battlefield, you gain life equal to the number
//	 of age counters on it."
//
// The cumulative upkeep is Yavimaya Ants' (CR 702.24a). The leaves trigger covers
// every exit, the sacrifice for an unpaid upkeep included, and counts the age
// counters the Unicorn had as it last existed (CR 603.10a).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "18a8b50b-2328-459b-979d-04c4c769e569",
		Name:         "Revered Unicorn",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.CumulativeUpkeep("Revered Unicorn — cumulative upkeep {1}", "{1}"),
			effects.WhenThisLeaves("Revered Unicorn — you gain life equal to the number of age counters on it", effects.S82OldB772GainLifeForAgeCounters),
		},
	})
}
