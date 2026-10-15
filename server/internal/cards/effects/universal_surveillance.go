package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Universal Surveillance — Sorcery {X}{U}{U}{U}: "Improvise ... Draw X
// cards."
//
// Improvise is Spec.TapCost; artifacts pay the generic {X} but never a
// {U}. No simplification.
func init() {
	Register(Spec{
		OracleID:     "c5a5c17d-d1e4-4d23-9d2b-1c2febdcb0e7",
		Name:         "Universal Surveillance",
		Completeness: CompletenessFull,
		XMatters:     true,
		TapCost:      Improvise(),
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			return DrawCards{Player: ctx.Controller(), N: ctx.X()}.Apply(ctx)
		},
	})
}
