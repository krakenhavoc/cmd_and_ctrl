package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Contact Other Plane — Instant {3}{U} (Adventures in the Forgotten
// Realms, #5499):
//
//	"Roll a d20.
//	 1—9 | Draw two cards.
//	 10—19 | Scry 2, then draw two cards.
//	 20 | Scry 3, then draw three cards."
//
// The draw after a scry waits for the scry to finish (Preordain's shape).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "a352cc67-eae8-4fa5-abba-507821c4cf47",
		Name:         "Contact Other Plane",
		Completeness: CompletenessFull,
		Purpose:      game.Purpose{Draws: 2},
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			return RollDice{Sides: 20, N: 1, Then: contactOtherPlaneTable}.Apply(ctx)
		},
	})
}

// contactOtherPlaneTable is the results table, after the d20.
var contactOtherPlaneTable = DiceTail("dice/contact-other-plane-table", func(ctx *Context, rolls []int) error {
	switch {
	case rolls[0] >= 20:
		return Scry{N: 3, ThenTail: ctx.TailN(thenDrawCards, 3)}.Apply(ctx)
	case rolls[0] >= 10:
		return Scry{N: 2, ThenTail: ctx.TailN(thenDrawCards, 2)}.Apply(ctx)
	}
	return DrawCards{N: 2}.Apply(ctx)
})
