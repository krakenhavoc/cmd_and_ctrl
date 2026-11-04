package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Pterafractyl — Creature — Dinosaur Fractal {X}{G}{U}, 1/0 (Secrets of
// Strixhaven, #4456):
//
//	"Flying
//	 This creature enters with X +1/+1 counters on it.
//	 When this creature enters, you gain 2 life."
//
// The X counters are the printed CR 614.1c entry clause (Walking Ballista's
// XCounters), so with X = 0 the 1/0 is put into the graveyard by state-based
// actions (CR 704.5f) and its enters trigger still resolves, gaining the 2 life.
// The gain is a plain enters trigger. Flying is a printed keyword. Not every
// part of the card scales with X (the life gain does not), so it is on the X
// guard's allowlist.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:                   "cca03b76-dd3c-4228-b97a-f7fa20819b2e",
		Name:                       "Pterafractyl",
		Completeness:               CompletenessFull,
		PrintedKeywords:            []string{"flying"},
		EntersWithCountersFromCast: []game.EntryCountersFromCast{XCounters(game.CounterPlusOne)},
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("Pterafractyl — you gain 2 life", Do(GainLife{Amount: 2})),
		},
	})
}
