package cardsw

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

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
	effects.Register(effects.Spec{
		OracleID: "54748cb1-d92a-4212-ad76-417ee79b5ef1",
		Name:     "Wash Out",
		// ADR 0126 §6: only permanents of the chosen color.
		Purpose:      game.Purpose{Sweep: game.Sweep{Matches: game.SweepAllPermanents, How: game.SweepBounce, Partial: true}},
		Completeness: effects.CompletenessFull,
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			effects.ChooseColorTail(game.ColorForHarm, ctx.Game, item.Controller, item.SourceCardID, "Wash Out — choose a color",
				ctx.ItemTail(washOutColorChosen))
			return nil
		},
	})
}

// washOutColorChosen returns every permanent of the colour (the tail's Word)
// to its owner's hand. An on-disk key.
var washOutColorChosen = effects.CardTail("then/wash-out-color-chosen", func(ctx *effects.Context, o game.TailOutcome) error {
	return effects.BounceAllMatching{Match: effects.OfColor(o.Tail.Word)}.Apply(ctx)
})
