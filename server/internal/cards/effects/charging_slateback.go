package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Charging Slateback — Creature — Beast {4}{R}, 4/3 (Onslaught):
//
//	"This creature can't block.
//	 Morph {4}{R}"
//
// The restriction is a static on the card, so face down (a 2/2 with no
// text, CR 708.2a) it can block; turned face up for {4}{R} it can't.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:         "67457ecb-5ff2-4f5d-9039-b9e80f54bba8",
		Name:             "Charging Slateback",
		Completeness:     CompletenessFull,
		Static:           []game.StaticAbility{RestrictSelf(game.CantBlock)},
		AlternativeCosts: []game.AlternativeCost{Morph("{4}{R}")},
	})
}
