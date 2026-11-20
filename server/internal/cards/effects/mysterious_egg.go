package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Mysterious Egg — Creature — Egg {1}, 0/2 (Ikoria: Lair of Behemoths, #5512):
//
//	"Whenever this creature mutates, put a +1/+1 counter on it."
//
// It has no mutate cost of its own: the trigger fires when a mutating
// creature spell merges with it, over or under (CR 702.140d), and the
// counter goes on the merged permanent (CR 730.2c).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "d3b62851-6013-49e5-8808-fa61b1bdfe98",
		Name:         "Mysterious Egg",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			On(game.EventMutate, ThisCreatureMutates, "Mysterious Egg — put a +1/+1 counter on it",
				Do(CounterOnThis{Kind: game.CounterPlusOne, N: 1})),
		},
	})
}
