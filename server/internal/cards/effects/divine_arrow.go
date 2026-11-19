package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Divine Arrow — Instant {1}{W} (Foundations Jumpstart, #5929):
//
//	"Divine Arrow deals 4 damage to target attacking or blocking creature."
//
// The target must be attacking or blocking when chosen and again as the
// spell resolves (CR 608.2b); a creature removed from combat in response
// takes nothing.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "6add505a-9c0f-42e6-898d-be3194c3df33",
		Name:         "Divine Arrow",
		Purpose:      ForTargets(DamageToTarget(0, 4)),
		Completeness: CompletenessFull,
		Targets:      TargetCreature("target attacking or blocking creature", AttackingOrBlocking()),
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			if len(item.Targets) == 0 || item.Targets[0].Kind != game.TargetCard {
				return nil
			}
			return DealDamage{Source: ctx.Source(), Target: item.Targets[0].ID, Amount: 4}.Apply(ctx)
		},
	})
}
