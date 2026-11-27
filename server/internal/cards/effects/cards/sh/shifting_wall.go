package cardssh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Shifting Wall — Artifact Creature — Wall {X}, 0/0 (Stronghold, S82 OLD batch
// 748, #5627):
//
//	"Defender (This creature can't attack.)
//	 This creature enters with X +1/+1 counters on it."
//
// Big Bertha's X counters (EntersWithCountersFromCast reads the X announced as
// the spell was cast, so a Wall put onto the battlefield some other way enters
// with none and dies as a 0/0). Defender is a printed keyword.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:                   "7c852dfe-8238-461e-814e-9667807f2cf5",
		Name:                       "Shifting Wall",
		Completeness:               effects.CompletenessFull,
		PrintedKeywords:            []string{"defender"},
		XMatters:                   true,
		EntersWithCountersFromCast: []game.EntryCountersFromCast{effects.XCounters(game.CounterPlusOne)},
	})
}
