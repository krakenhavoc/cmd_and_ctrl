package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Locust Miser — Creature — Rat Shaman {2}{B}{B}, 2/2:
//
//	"Each opponent's maximum hand size is reduced by two."
//
// Gnat Miser's bigger sibling: a maximum-hand-size static (ADR 0113 §3,
// #2074), folded in CR 613.11 timestamp order.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "75a53a88-51a5-4c76-9f67-ee94c3065b98",
		Name:         "Locust Miser",
		Completeness: CompletenessFull,
		HandSize:     []game.HandSizeStatic{EachOpponentsMaxHandSizeReducedBy(2)},
	})
}
