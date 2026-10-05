package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Distant Melody — Sorcery {3}{U}:
//
//	"Choose a creature type. Draw a card for each permanent you
//	 control of that type."
//
// The type is chosen as the spell resolves (#2382); the draw count is
// read when the answer arrives, from the board as it is then. It counts
// PERMANENTS, not creatures, so a Kindred artifact or a changeling
// counts, and a Zombie you control that is not a creature does too.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "ac6c5852-71c7-4f19-8c87-ccb345052862",
		Name:         "Distant Melody",
		Completeness: CompletenessFull,
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			ChooseCreatureTypeThen(ctx.Game, item.Controller, item.SourceCardID, "Distant Melody — choose a creature type",
				func(g *game.Game, t string) error {
					if t == "" {
						return nil
					}
					return DrawCards{Player: item.Controller, N: permanentsOfTypeYouControl(g, item.Controller, t)}.Apply(NewContext(g, item))
				})
			return nil
		},
	})
}
