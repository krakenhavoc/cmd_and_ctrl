package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Hallowed Burial — Sorcery {3}{W}{W} (Eventide, #5590):
//
//	"Put all creatures on the bottom of their owners' libraries."
//
// Terminus without the miracle: every creature goes at once, and each
// owner with more than one orders their own (CR 401.4).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "dcaaddee-00b1-495c-9161-40c91c900aac",
		Name:         "Hallowed Burial",
		Completeness: CompletenessFull,
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			return tuckAllCreaturesToBottomByOwner(ctx, item, "Hallowed Burial — put these on the bottom of your library in any order")
		},
	})
}
