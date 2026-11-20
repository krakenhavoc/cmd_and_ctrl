package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Refurbish — Sorcery {3}{W} (Kaladesh, #5539):
//
//	"Return target artifact card from your graveyard to the
//	 battlefield."
//
// Argivian Restoration's body: the artifact card returns under its
// caster's control.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "e4e61fe3-27a4-4149-bdfb-67dcfc8eb32c",
		Name:         "Refurbish",
		Completeness: CompletenessFull,
		Targets:      TargetCardInGraveyard("target artifact card from your graveyard", Artifact(), YouOwn()),
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			_, _ = reanimateSingleTarget(ctx, ctx.Controller())
			return nil
		},
	})
}
