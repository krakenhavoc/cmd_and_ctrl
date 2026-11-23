package cardss

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Seismic Sense — Sorcery — Lesson {G} (Avatar: The Last Airbender, #4459):
//
//	"Look at the top X cards of your library, where X is the number of
//	 lands you control. You may reveal a creature or land card from among
//	 them and put it into your hand. Put the rest on the bottom of your
//	 library in a random order."
//
// Horn of the Mark's sentence (LookAtTopThenMayTakeToHand) with X counted
// as the spell resolves: the lands its controller controls then, read
// through the layers. No lands looks at nothing.
//
// No simplification.
func init() {
	const label = "Seismic Sense — reveal a creature or land card and put it into your hand"
	effects.Register(effects.Spec{
		OracleID:     "9d586c74-b1dd-4fcb-a24a-c703d250fa63",
		Name:         "Seismic Sense",
		Completeness: effects.CompletenessFull,
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			x := effects.B10LandsControlled(ctx.Game, ctx.Controller())
			if x <= 0 {
				return nil
			}
			return effects.LookAtTopThenMayTakeToHand(x, effects.Or(effects.Creature(), effects.Land()), 1, label)(ctx.Game, item)
		},
	})
}
