package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	Register(Spec{
		OracleID:     "e59c07ab-08fd-4d5b-978e-1594604ba410",
		Name:         "Ultimate Alliance",
		Completeness: CompletenessFull,
		Targets:      TargetCreature("target creature"),
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			for _, t := range ctx.LegalTargets() {
				return DealDamage{
					Source: item.SourceCardID,
					Target: t.ID,
					Amount: b04CreaturesControlled(ctx.Game, ctx.Controller()),
				}.Apply(ctx)
			}
			return nil
		},
	})
}
