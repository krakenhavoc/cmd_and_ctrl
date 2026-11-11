package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Hooded Hydra — Creature — Snake Hydra {X}{G}{G}, 0/0 (Murders at
// Karlov Manor Commander):
//
//	"This creature enters with X +1/+1 counters on it.
//	 When this creature dies, create a 1/1 green Snake creature token for
//	 each +1/+1 counter on it.
//	 Morph {3}{G}{G}
//	 As this creature is turned face up, put five +1/+1 counters on it."
//
// Cast face up it enters with X counters (Feral Hydra's XCounters). Cast
// face down for {3} there is no X and it has no abilities (CR 708.2a), so
// it enters a bare 2/2; turning it up puts five counters on it as a
// CR 614.1e replacement, before anything sees it (CR 708.11). The
// Snakes are counted from the Hydra's last-known information when it
// dies (CR 603.10, CR 608.2h): Ribtruss Roaster's count, counters of
// every source included.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:                   "637ddd65-112d-40f7-9bab-a61724d9165c",
		Name:                       "Hooded Hydra",
		Completeness:               CompletenessFull,
		XMatters:                   true,
		EntersWithCountersFromCast: []game.EntryCountersFromCast{XCounters(game.CounterPlusOne)},
		AlternativeCosts: []game.AlternativeCost{
			AsTurnedFaceUp(Morph("{3}{G}{G}"), game.AsTurnedFaceUpCounters{N: 5}),
		},
		Triggered: []game.TriggeredAbility{
			WhenThisDies("Hooded Hydra — create a 1/1 green Snake for each +1/+1 counter on it",
				Do(CreateTokensForEach{
					Template: TokenCard("1/1 green Snake"),
					Name:     "the +1/+1 counters on it",
					Count:    socB9PlusOneCountersOnSource,
				})),
		},
	})
}
