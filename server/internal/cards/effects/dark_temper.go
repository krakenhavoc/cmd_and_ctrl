package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Dark Temper — Instant {2}{R} (Conflux, #5587):
//
//	"Dark Temper deals 2 damage to target creature. If you control a
//	 black permanent, destroy the creature instead."
//
// Which half happens is read as the spell resolves (CR 608.2h). With a
// black permanent the creature is destroyed and no damage is dealt; a
// target that left in response gets nothing (CR 608.2b).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "3a0ea41e-b4f5-4aa1-9f7a-158485313c64",
		Name:         "Dark Temper",
		Completeness: CompletenessFull,
		Targets:      TargetCreature("target creature"),
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			id, ok := b16FirstLegalTargetCard(ctx)
			if !ok {
				return nil
			}
			if b18ControlsPermanentOfColor(ctx.Game, item.Controller, ctx.TextColor("B")) {
				return DestroyTarget{Target: id}.Apply(ctx)
			}
			return DealDamage{Source: item.SourceCardID, Target: id, Amount: 2}.Apply(ctx)
		},
	})
}
