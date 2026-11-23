package cardsa

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Ainok Artillerist — Creature — Dog Archer {2}{G}, 4/1 (Foundations
// Jumpstart):
//
//	"This creature has reach as long as it has a +1/+1 counter on it."
//
// Chaos Imps' self grant read off the live counters.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "860d0449-8032-458b-8335-30be238d6689",
		Name:         "Ainok Artillerist",
		Completeness: effects.CompletenessFull,
		Static:       []game.StaticAbility{effects.ThisHasWhileItHasCounter(game.CounterPlusOne, "reach")},
	})
}
