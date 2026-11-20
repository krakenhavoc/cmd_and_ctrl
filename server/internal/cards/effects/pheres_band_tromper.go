package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Pheres-Band Tromper — Creature — Centaur Warrior {3}{G}, 3/3 (Born of the Gods,
// #5559):
//
//	"Inspired — Whenever this creature becomes untapped, put a +1/+1
//	 counter on it."
//
// Inspired is an ability word: the trigger is Key to the City's
// (EventUntapCard on itself), so it fires in the untap step and on
// any other untap.
// The counter is Akroan Skyguard's (thisGetsAPlusOneCounter).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "41b2876f-0217-4552-9fad-f12f1824d06f",
		Name:         "Pheres-Band Tromper",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			On(game.EventUntapCard, Self, "Pheres-Band Tromper — a +1/+1 counter on it (inspired)", thisGetsAPlusOneCounter),
		},
	})
}
