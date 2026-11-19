package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Ainok Artillerist — Creature — Dog Archer {2}{G}, 4/1 (Foundations
// Jumpstart):
//
//	"This creature has reach as long as it has a +1/+1 counter on it."
//
// Chaos Imps' self grant read off the live counters.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "860d0449-8032-458b-8335-30be238d6689",
		Name:         "Ainok Artillerist",
		Completeness: CompletenessFull,
		Static:       []game.StaticAbility{ThisHasWhileItHasCounter(game.CounterPlusOne, "reach")},
	})
}
