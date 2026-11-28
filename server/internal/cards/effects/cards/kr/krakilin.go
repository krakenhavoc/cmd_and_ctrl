package cardskr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Krakilin — Creature — Beast {X}{G}{G}, 0/0 (Tempest, S82 OLD batch 759,
// #5628):
//
//	"This creature enters with X +1/+1 counters on it.
//	 {1}{G}: Regenerate this creature."
//
// Benevolent Hydra's X counters from the cast, and Spiritmonger's regenerate
// row. X is the announced X of the spell, so a Krakilin put onto the
// battlefield another way has X = 0 and dies as a 0/0 (CR 107.3b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:                   "3eb49097-42f8-4388-907d-32a738749d8a",
		Name:                       "Krakilin",
		Completeness:               effects.CompletenessFull,
		XMatters:                   true,
		EntersWithCountersFromCast: []game.EntryCountersFromCast{effects.XCounters(game.CounterPlusOne)},
		Activated:                  []effects.ActivatedAbility{effects.RegenerateThisRow("{1}{G}")},
	})
}
