package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Argivian Restoration — Sorcery {2}{U}{U} (Weatherlight, #5629):
//
//	"Return target artifact card from your graveyard to the battlefield."
//
// Zombify's return for an artifact card: it enters under the caster's
// control. A target that left the graveyard in response does nothing
// (CR 608.2b).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "d887919a-68d5-455d-977c-4f1f0b4b3141",
		Name:         "Argivian Restoration",
		Completeness: CompletenessFull,
		Targets:      TargetCardInGraveyard("target artifact card from your graveyard", Artifact(), YouOwn()),
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			_, _ = reanimateSingleTarget(ctx, ctx.Controller())
			return nil
		},
	})
}
