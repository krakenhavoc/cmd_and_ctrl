package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Woodland Wanderer — Creature — Elemental {3}{G}, 2/2 (Battle for
// Zendikar, #5547):
//
//	"Vigilance, trample
//	 Converge — This creature enters with a +1/+1 counter on it for
//	 each color of mana spent to cast it."
//
// Vigilance and trample are carried as printed data. Converge counts
// the distinct colours of mana spent to cast it, read off the payment
// the permanent carries (Crystalline Crawler's seam): cast for
// colourless mana, or not cast at all, it enters with no counters.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:            "d5f3c711-edfa-4105-9844-e3f444942a03",
		Name:                "Woodland Wanderer",
		Completeness:        CompletenessFull,
		PrintedKeywords:     []string{"vigilance", "trample"},
		ReadsManaSpent:      game.ReadsOwnPayment,
		WantsDistinctColors: true,
		EntersWithCountersFromCast: []game.EntryCountersFromCast{
			CountersPerColorSpent(game.CounterPlusOne),
		},
	})
}
