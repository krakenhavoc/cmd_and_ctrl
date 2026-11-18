package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Wanderer's Intervention — Instant {1}{W} (Kamigawa: Neon Dynasty, #5493):
//
//	"Wanderer's Intervention deals 4 damage to target attacking or
//	 blocking creature."
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "4497d0b9-cf4b-4c9d-bf7e-8ed85aa6b5b8",
		Name:         "Wanderer's Intervention",
		Completeness: CompletenessFull,
		Purpose:      ForTargets(DamageToTarget(0, 4)),
		Targets:      TargetCreature("target attacking or blocking creature", AttackingOrBlocking()),
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			if len(item.Targets) == 0 || !ctx.IsTargetLegal(item.Targets[0]) {
				return nil
			}
			return DealDamage{Source: ctx.Source(), Target: item.Targets[0].ID, Amount: 4}.Apply(ctx)
		},
	})
}
