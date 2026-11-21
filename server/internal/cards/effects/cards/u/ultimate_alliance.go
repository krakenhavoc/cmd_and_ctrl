package cardsu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Ultimate Alliance — Instant {W} (Marvel Super Heroes Commander, #4455):
//
//	"Ultimate Alliance deals damage equal to the number of creatures you
//	 control to target creature."
//
// The count is taken as the spell resolves, so a creature that died in
// response is not counted; the target is dealt damage only if it is still
// legal.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "e59c07ab-08fd-4d5b-978e-1594604ba410",
		Name:         "Ultimate Alliance",
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetCreature("target creature"),
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			for _, t := range ctx.LegalTargets() {
				return effects.DealDamage{
					Source: item.SourceCardID,
					Target: t.ID,
					Amount: effects.B04CreaturesControlled(ctx.Game, ctx.Controller()),
				}.Apply(ctx)
			}
			return nil
		},
	})
}
