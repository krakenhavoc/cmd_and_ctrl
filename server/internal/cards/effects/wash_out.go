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
		OracleID: "54748cb1-d92a-4212-ad76-417ee79b5ef1",
		Name:     "Wash Out",
		// ADR 0126 §6: only permanents of the chosen color.
		Purpose:      game.Purpose{Sweep: game.Sweep{Matches: game.SweepAllPermanents, How: game.SweepBounce, Partial: true}},
		Completeness: CompletenessFull,
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			ChooseColorTail(game.ColorForHarm, ctx.Game, item.Controller, item.SourceCardID, "Wash Out — choose a color",
				ctx.ItemTail(washOutColorChosen))
			return nil
		},
	})
}

// washOutColorChosen returns every permanent of the colour (the tail's Word)
// to its owner's hand. An on-disk key.
var washOutColorChosen = CardTail("then/wash-out-color-chosen", func(ctx *Context, o game.TailOutcome) error {
	return BounceAllMatching{Match: OfColor(o.Tail.Word)}.Apply(ctx)
})
