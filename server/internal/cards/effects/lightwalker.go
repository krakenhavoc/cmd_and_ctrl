package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Lightwalker — Creature — Human Warrior {1}{W}, 2/1:
//
//	"This creature has flying as long as it has a +1/+1 counter on it."
//
// Chaos Imps' layer-6 self grant, read off the live counters.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "6f4848d4-a5dc-4c71-955c-ad2b1cbb7ede",
		Name:         "Lightwalker",
		Completeness: CompletenessFull,
		Static:       []game.StaticAbility{ThisHasWhileItHasCounter(game.CounterPlusOne, "flying")},
	})
}
