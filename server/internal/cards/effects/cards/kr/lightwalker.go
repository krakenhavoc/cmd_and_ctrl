package cardskr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Lightwalker — Creature — Human Warrior {1}{W}, 2/1:
//
//	"This creature has flying as long as it has a +1/+1 counter on it."
//
// Chaos Imps' layer-6 self grant, read off the live counters.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "6f4848d4-a5dc-4c71-955c-ad2b1cbb7ede",
		Name:         "Lightwalker",
		Completeness: effects.CompletenessFull,
		Static:       []game.StaticAbility{effects.ThisHasWhileItHasCounter(game.CounterPlusOne, "flying")},
	})
}
