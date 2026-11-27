package cardsb

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Benalish Lancer — Creature — Human Knight {2}{W}, 2/2 (Invasion, S82 OLD
// #5616):
//
//	"Kicker {2}{W} (You may pay an additional {2}{W} as you cast this spell.)
//	 If this creature was kicked, it enters with two +1/+1 counters on it and
//	 with first strike."
//
// The counters are the kicked entry (CR 614.1c). First strike is a layer 6
// grant keyed on the permanent's own kicker record (game.CardKickedTimes,
// CR 400.7d), as the Volvers' keywords are, so a Lancer that was never cast
// has neither.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:                   "b5adbd37-8d98-4c65-b42d-cd9db94e7222",
		Name:                       "Benalish Lancer",
		Completeness:               effects.CompletenessFull,
		OptionalCosts:              []game.AdditionalCost{effects.Kicker("{2}{W}")},
		EntersWithCountersFromCast: []game.EntryCountersFromCast{effects.CountersPerKick(game.CounterPlusOne, 2)},
		Static:                     []game.StaticAbility{effects.OldB650KickedKeywordGrant("first strike")},
	})
}
