package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	Register(Spec{
		OracleID:     "90aa9113-5148-486b-a1c6-b4230e6d3e41",
		Name:         "Unwanted Remake",
		Completeness: CompletenessFull,
		Purpose:      game.Purpose{Answers: game.AnswerRemove},
		Targets:      TargetCreature("target creature"),
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			t, ok := ctx.ClauseTarget(0)
			if !ok {
				return nil
			}
			controller, ok := controllerOfTarget(ctx, t.ID)
			if !ok {
				return nil
			}
			if err := (DestroyTarget{Target: t.ID}).Apply(ctx); err != nil {
				return err
			}
			return ManifestDread{Player: controller}.Apply(ctx)
		},
	})
}
