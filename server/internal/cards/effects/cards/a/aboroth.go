package cardsa

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Aboroth — {4}{G}{G} Creature — Elemental, 9/9 (Weatherlight, S82 OLD batch 772,
// #5629):
//
//	"Cumulative upkeep—Put a -1/-1 counter on this creature."
//
// The cumulative upkeep's payment is the counters themselves (#6189,
// CountersOnThisPayment): with N age counters it puts N -1/-1 counters on the
// Aboroth, all at once or none (CR 702.24a), or the Aboroth is sacrificed. The
// permanent the counters go on is the asking one, read as the trigger resolves.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "28f70f86-19a9-4811-bc10-423a05842d39",
		Name:         "Aboroth",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.CumulativeUpkeepPaying("Aboroth — cumulative upkeep: put a -1/-1 counter on this creature",
				effects.CountersOnThisPayment(1, game.CounterMinusOne, uuid.Nil), "", ""),
		},
	})
}
