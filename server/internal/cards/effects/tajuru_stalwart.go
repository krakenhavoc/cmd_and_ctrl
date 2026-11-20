package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Tajuru Stalwart — Creature — Elf Scout Ally {2}{G}, 0/1 (Battle for
// Zendikar, #5547):
//
//	"Converge — This creature enters with a +1/+1 counter on it for
//	 each color of mana spent to cast it."
//
// Converge counts the distinct colours of mana spent to cast it, read
// off the payment the permanent carries (Crystalline Crawler's seam):
// cast for colourless mana, or not cast at all, it enters with no
// counters.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:            "ac87cc33-4a71-471b-917d-6a92721ea0b9",
		Name:                "Tajuru Stalwart",
		Completeness:        CompletenessFull,
		ReadsManaSpent:      game.ReadsOwnPayment,
		WantsDistinctColors: true,
		EntersWithCountersFromCast: []game.EntryCountersFromCast{
			CountersPerColorSpent(game.CounterPlusOne),
		},
	})
}
