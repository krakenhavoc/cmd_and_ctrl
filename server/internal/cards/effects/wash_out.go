package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Wash Out — Sorcery {3}{U}:
//
//	"Return all permanents of the color of your choice to their owners'
//	 hands."
//
// The colour is chosen as the spell resolves, through #742's
// resolution-time prompt; the continuation then bounces every
// permanent of that colour as one simultaneous event
// (BounceAllMatching). A multicoloured permanent is "of" each of its
// colours and goes; a colourless one never does. Tokens cease to exist
// in hand as usual.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "54748cb1-d92a-4212-ad76-417ee79b5ef1",
		Name:         "Wash Out",
		Completeness: CompletenessFull,
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			ChooseColorThen(game.ColorForHarm, ctx.Game, item.Controller, item.SourceCardID, "Wash Out — choose a color",
				func(g *game.Game, color string) error {
					return BounceAllMatching{Match: OfColor(color)}.Apply(NewContext(g, item))
				})
			return nil
		},
	})
}
