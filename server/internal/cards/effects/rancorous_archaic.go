package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Rancorous Archaic — Creature — Avatar {5}, 2/2 (Secrets of Strixhaven,
// #4456):
//
//	"Reach, trample
//	 Converge — This creature enters with a +1/+1 counter on it for each
//	 color of mana spent to cast it."
//
// Converge is an "enters with" clause read off its own payment
// (CountersPerColorSpent, Wildgrowth Archaic's): five mana of three colours
// is three counters, and colourless mana adds none. The cost is all generic,
// so WantsDistinctColors steers the auto-tapper to pay with as many colours
// as the pool allows, and ReadsManaSpent makes a waived payment ask what was
// paid.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:            "6a68fcaf-b1e8-451d-8e1c-b47db55c83cc",
		Name:                "Rancorous Archaic",
		ReadsManaSpent:      game.ReadsOwnPayment,
		Completeness:        CompletenessFull,
		WantsDistinctColors: true,
		PrintedKeywords:     []string{"reach", "trample"},
		EntersWithCountersFromCast: []game.EntryCountersFromCast{
			CountersPerColorSpent(game.CounterPlusOne),
		},
	})
}
