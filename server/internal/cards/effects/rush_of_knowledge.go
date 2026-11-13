package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Rush of Knowledge — Sorcery {4}{U} (Scourge):
//
//	"Draw cards equal to the greatest mana value among permanents you
//	 control."
//
// The count is read as the spell resolves (scgB1GreatestManaValue), so the
// spell itself, on the stack, is not among the permanents. A board of
// lands alone (mana value 0) draws nothing.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "34282c99-68ef-4855-b683-7bb205fa6374",
		Name:         "Rush of Knowledge",
		Completeness: CompletenessFull,
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			n := scgB1GreatestManaValue(ctx)
			if n <= 0 {
				return nil
			}
			return DrawCards{Player: ctx.Controller(), N: n}.Apply(ctx)
		},
	})
}
