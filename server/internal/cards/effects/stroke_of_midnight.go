package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Stroke of Midnight — Instant {2}{W} (EDHREC rank 205):
//
//	"Destroy target nonland permanent. Its controller creates a 1/1
//	 white Human creature token."
//
// Generous Gift that can't hit a land and hands back a 1/1 instead
// of a 3/3 — a smaller drawback for a smaller target set. The shape
// is Beast Within's exactly: read the controller, destroy, then
// create under the victim. Nonland() is the whole difference in the
// target clause.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "9a107e48-3d50-4941-95b1-10f2b29a4245",
		Name:         "Stroke of Midnight",
		Completeness: CompletenessFull,
		Targets:      TargetPermanent("target nonland permanent", Nonland()),
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			if len(item.Targets) == 0 || item.Targets[0].Kind != game.TargetCard {
				return nil
			}
			target := item.Targets[0].ID
			controller, ok := controllerOfTarget(ctx, target)
			if !ok {
				return nil
			}
			if err := (DestroyTarget{Target: target}).Apply(ctx); err != nil {
				return err
			}
			return CreateToken{
				Controller: controller,
				Template:   TokenCard("1/1 colorless Human"),
				N:          1,
			}.Apply(ctx)
		},
	})
}
