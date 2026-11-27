package cardsn

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Nimbus Swimmer — Creature — Leviathan {X}{G}{U}, 0/0 (Gatecrash, S82 OLD
// #5565):
//
//	"Flying
//	 This creature enters with X +1/+1 counters on it."
//
// Benevolent Hydra's entry clause: X +1/+1 counters ride the CR 614.1c entry
// replacement as the X announced at the cast (XCounters), so with X = 0 it is a
// 0/0 that dies to state-based actions.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:                   "ce02ef24-d53e-4025-bc45-ca27aba89fe9",
		Name:                       "Nimbus Swimmer",
		Completeness:               effects.CompletenessFull,
		PrintedKeywords:            []string{"flying"},
		XMatters:                   true,
		EntersWithCountersFromCast: []game.EntryCountersFromCast{effects.XCounters(game.CounterPlusOne)},
	})
}
