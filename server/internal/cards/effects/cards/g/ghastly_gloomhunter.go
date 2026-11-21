package cardsg

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Ghastly Gloomhunter — Creature — Zombie Bat {1}{B}, 1/1 (Zendikar
// Rising, #5508):
//
//	"Kicker {3}{B} (You may pay an additional {3}{B} as you cast this
//	 spell.)
//	 Flying, lifelink
//	 If this creature was kicked, it enters with two +1/+1 counters on
//	 it."
//
// Skyclave Shade's kicked entry (CountersPerKick, a CR 614.1c clause read
// off the cast) on a printed flying, lifelink body.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:                   "70ab65ff-724f-4126-b472-97fa03ca3a34",
		Name:                       "Ghastly Gloomhunter",
		Completeness:               effects.CompletenessFull,
		PrintedKeywords:            []string{"flying", "lifelink"},
		OptionalCosts:              []game.AdditionalCost{effects.Kicker("{3}{B}")},
		EntersWithCountersFromCast: []game.EntryCountersFromCast{effects.CountersPerKick(game.CounterPlusOne, 2)},
	})
}
