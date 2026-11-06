package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	Register(Spec{
		OracleID:     "9d586c74-b1dd-4fcb-a24a-c703d250fa63",
		Name:         "Seismic Sense",
		Completeness: CompletenessFull,
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			x := b10LandsControlled(ctx.Game, ctx.Controller())
			if x <= 0 {
				return nil
			}
			return LookAtTopThenMayTakeToHand(x, Or(Creature(), Land()), 1, label)(ctx.Game, item)
		},
	})
}
