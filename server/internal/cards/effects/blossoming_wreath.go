package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Blossoming Wreath — Instant {G} (Weatherlight, #5629):
//
//	"You gain life equal to the number of creature cards in your graveyard."
//
// The creature cards are counted once, as the spell resolves
// (CR 608.2h).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "fbd99a88-8926-455e-8e23-58fed6d14719",
		Name:         "Blossoming Wreath",
		Completeness: CompletenessFull,
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			return GainLife{Amount: b11CreatureCardsInGraveyard(ctx.Game, ctx.Controller())}.Apply(ctx)
		},
	})
}
