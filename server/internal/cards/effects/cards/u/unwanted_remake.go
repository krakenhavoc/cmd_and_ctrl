package cardsu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Unwanted Remake — Instant {W} (Duskmourn: House of Horror, #4466):
//
//	"Destroy target creature. Its controller manifests dread."
//
// Fateful Absence's shape: the controller is read before the destroy,
// then that player is handed the compensation, here a manifest dread
// (CR 701.62a) they make themselves. The manifest does not depend on
// the destroy: an indestructible or regenerated creature stays, and its
// controller still manifests dread. A target that is gone fizzles the
// whole spell (CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "90aa9113-5148-486b-a1c6-b4230e6d3e41",
		Name:         "Unwanted Remake",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Answers: game.AnswerRemove},
		Targets:      effects.TargetCreature("target creature"),
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			t, ok := ctx.ClauseTarget(0)
			if !ok {
				return nil
			}
			controller, ok := effects.ControllerOfTarget(ctx, t.ID)
			if !ok {
				return nil
			}
			if err := (effects.DestroyTarget{Target: t.ID}).Apply(ctx); err != nil {
				return err
			}
			return effects.ManifestDread{Player: controller}.Apply(ctx)
		},
	})
}
