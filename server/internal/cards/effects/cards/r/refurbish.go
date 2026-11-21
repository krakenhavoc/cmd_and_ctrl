package cardsr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

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
	effects.Register(effects.Spec{
		OracleID:     "e4e61fe3-27a4-4149-bdfb-67dcfc8eb32c",
		Name:         "Refurbish",
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetCardInGraveyard("target artifact card from your graveyard", effects.Artifact(), effects.YouOwn()),
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			_, _ = effects.ReanimateSingleTarget(ctx, ctx.Controller())
			return nil
		},
	})
}
