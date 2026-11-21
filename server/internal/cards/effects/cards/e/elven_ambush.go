package cardse

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Elven Ambush — Instant {3}{G} (EDHREC rank 2798):
//
//	"Create a 1/1 green Elf Warrior creature token for each Elf you
//	 control."
//
// The Elf deck's instant-speed doubler. "Each Elf you control" is
// every permanent with the type — a changeling counts — read as the
// spell resolves. The tokens are Elves, so a second Ambush makes
// twice as many.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "a2373025-20c1-4416-91f7-e51f68dbd146",
		Name:         "Elven Ambush",
		Completeness: effects.CompletenessFull,
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			n := effects.B26PermanentsOfSubtypeControlled(ctx.Game, ctx.Controller(), ctx.TextSubtype("Elf"))
			if n <= 0 {
				return nil
			}
			return effects.CreateToken{Controller: ctx.Controller(), Template: effects.TokenCard("1/1 green Elf Warrior"), N: n}.Apply(ctx)
		},
	})
}
