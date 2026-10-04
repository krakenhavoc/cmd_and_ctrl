package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Gnat Miser — Creature — Rat Shaman {B}, 1/1:
//
//	"Each opponent's maximum hand size is reduced by one."
//
// A maximum-hand-size static (ADR 0113 §3, #2074), folded in CR 613.11
// timestamp order with every other one that reaches the player.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "cfb5da72-9f00-4ae2-8d59-f997faab451b",
		Name:         "Gnat Miser",
		Completeness: CompletenessFull,
		HandSize:     []game.HandSizeStatic{EachOpponentsMaxHandSizeReducedBy(1)},
	})
}
