package cardsw

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Worldsoul Colossus — Creature — Elemental {X}{G}{W}, 0/0 (Guilds of
// Ravnica):
//
//	"Convoke
//	 This creature enters with X +1/+1 counters on it."
//
// Chord of Calling's convoke over an {X} cost and Big Bertha's X counters: the
// announced X is the number of counters the creature enters with (CR 614.1c).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:                   "6970f285-dc2b-437b-b099-ca1d3ef106ab",
		Name:                       "Worldsoul Colossus",
		Completeness:               effects.CompletenessFull,
		XMatters:                   true,
		TapCost:                    effects.Convoke(),
		EntersWithCountersFromCast: []game.EntryCountersFromCast{effects.XCounters(game.CounterPlusOne)},
	})
}
