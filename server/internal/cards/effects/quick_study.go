package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Quick Study — Instant {2}{U} (EDHREC rank 2644):
//
//	"Draw two cards."
//
// Divination at instant speed — the cantrip-and-a-half every blue
// deck holds up on an opponent's end step. One primitive.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "7bff8d4a-1c7d-48b8-b3e3-737dd6f01823",
		Name:         "Quick Study",
		Completeness: CompletenessFull,
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			return DrawCards{Player: item.Controller, N: 2}.Apply(ctx)
		},
	})
}
