package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Neverwinter Hydra — Creature — Hydra {X}{X}{G}{G}, 0/0 (Forgotten
// Realms Commander, #5682):
//
//	"As this creature enters, roll X d6. It enters with a number of
//	 +1/+1 counters on it equal to the total of those results.
//	 Trample
//	 Ward {4}"
//
// The entry clause is RollXDiceCounters: X is the spell's X, the dice
// are rolled as the spell resolves and before the Hydra enters, and the
// total rides the entry event as its counters, so Doubling Season and
// Hardened Scales apply and the Hydra is never on the battlefield as a
// 0/0. Reanimated, blinked or cast for X=0 it rolls nothing and enters
// with no counters (CR 107.3m), and the next state-based check puts it
// into its owner's graveyard.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:                   "e71e7a68-d96e-4ebb-8d4f-9c45ff8b8b66",
		Name:                       "Neverwinter Hydra",
		Completeness:               CompletenessFull,
		XMatters:                   true,
		PrintedKeywords:            []string{"trample"},
		EntersWithCountersFromCast: []game.EntryCountersFromCast{RollXDiceCounters(game.CounterPlusOne, 6)},
		Triggered: []game.TriggeredAbility{
			Ward(WardMana("{4}"), "Neverwinter Hydra — ward {4}"),
		},
	})
}
