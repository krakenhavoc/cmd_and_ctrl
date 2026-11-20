package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Mystic Speculation — Sorcery {U} (Future Sight, #5594):
//
//	"Buyback {2} (You may pay an additional {2} as you cast this spell.
//	 If you do, put this card into your hand as it resolves.)
//	 Scry 3."
//
// A scry 3 with buyback (CR 702.27): the optional {2} is the shared
// Buyback cost, and the return to hand is the engine's.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:      "a58adf82-d9e2-4b47-933e-68180f280ba2",
		Name:          "Mystic Speculation",
		Completeness:  CompletenessFull,
		Purpose:       game.Purpose{Scry: 3},
		OptionalCosts: []game.AdditionalCost{Buyback("{2}")},
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			return Scry{Player: ctx.Controller(), N: 3}.Apply(ctx)
		},
	})
}
