package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Endless One — Creature — Eldrazi {X}, 0/0 (Battle for Zendikar,
// #5547):
//
//	"This creature enters with X +1/+1 counters on it."
//
// Shivan Devastator's entry: X is announced with the cast (CR 601.2b)
// and the counters are part of the entry (CR 614.1c), so it never
// exists as a 0/0. Put onto the battlefield without being cast, X is 0
// and it dies to the state-based check (CR 107.3m, CR 704.5f).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:                   "7a51780a-fa28-4ee0-94c7-4330800ca9cb",
		Name:                       "Endless One",
		Completeness:               CompletenessFull,
		XMatters:                   true,
		EntersWithCountersFromCast: []game.EntryCountersFromCast{XCounters(game.CounterPlusOne)},
	})
}
