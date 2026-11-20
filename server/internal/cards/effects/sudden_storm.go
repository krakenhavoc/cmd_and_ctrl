package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Sudden Storm — Instant {3}{U} (Born of the Gods,
// #5559):
//
//	"Tap up to two target creatures. Those creatures don't untap during
//	 their controllers' next untap steps. Scry 1. (Look at the top card of
//	 your library. You may put that card on the bottom.)"
//
// Frost Breath's tap-and-freeze (b751TapFreezeTargets), then the scry.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "32e915fb-3836-4bed-93b1-61c9c3951ad1",
		Name:         "Sudden Storm",
		Completeness: CompletenessFull,
		Purpose:      game.Purpose{Scry: 1},
		Targets:      TargetCreature("up to two target creatures").WithCount(0, 2),
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			if err := b751TapFreezeTargets(ctx, uuid.Nil, "Sudden Storm"); err != nil {
				return err
			}
			return Scry{Player: ctx.Controller(), N: 1}.Apply(ctx)
		},
	})
}
