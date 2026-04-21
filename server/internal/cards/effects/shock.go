package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Shock — "Shock deals 2 damage to any target." Same shape as
// Lightning Bolt with a different constant; proves the primitive
// is reused across cards.
func init() {
	Register(Spec{
		OracleID:   "a9d288b8-cdc1-4e55-a0c9-d6edfc95e65d",
		Name:       "Shock",
		TargetMode: "any",
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			if len(item.Targets) == 0 {
				return nil
			}
			return DealDamage{
				Source: ctx.Source(),
				Target: item.Targets[0].ID,
				Amount: 2,
			}.Apply(ctx)
		},
	})
}
