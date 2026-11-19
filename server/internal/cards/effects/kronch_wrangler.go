package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Kronch Wrangler — Creature — Human Warrior {1}{G}, 2/1 (Ravnica: Clue
// Edition, Ravnica Remastered):
//
//	"Trample
//	 Whenever a creature you control with power 4 or greater enters, put a
//	 +1/+1 counter on this creature."
//
// The entering creature's power is read as it entered (Garruk, Curse
// Breaker's shape), counters it entered with included. The Wrangler's own
// entry never counts: it is a 2/1. The counter goes on the Wrangler only if
// it is still the same permanent when the trigger resolves (CR 400.7).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "635dda8c-3761-4c87-8723-8cd09878516b",
		Name:            "Kronch Wrangler",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"trample"},
		Triggered: []game.TriggeredAbility{
			On(game.EventETB, garrukCurseBreakerBigEntered,
				"Kronch Wrangler — put a +1/+1 counter on this creature",
				Do(CounterOnThis{Kind: game.CounterPlusOne, N: 1})),
		},
	})
}
