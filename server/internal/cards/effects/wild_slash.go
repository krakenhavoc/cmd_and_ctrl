package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Wild Slash — Instant {R}:
//
//	"Ferocious — If you control a creature with power 4 or greater,
//	 damage can't be prevented this turn.
//	 Wild Slash deals 2 damage to any target."
//
// Ferocious is checked as the spell resolves, before the damage, and
// gates only the turn grant (ADR 0107 §5): the 2 damage happens either
// way.
//
// No simplifications.
func init() {
	Register(Spec{
		OracleID:     "f76a0e5b-74c1-4ee8-b502-e8b988086de4",
		Name:         "Wild Slash",
		Completeness: CompletenessFull,
		Targets:      TargetAny(),
		Purpose:      ForTargets(DamageToTarget(0, 2)),
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			if b24ControlsCreatureWithPowerAtLeast(ctx.Game, ctx.Controller(), 4) {
				if err := (DamageCantBePreventedThisTurn{}).Apply(ctx); err != nil {
					return err
				}
			}
			return damageToFirstTarget(2)(item, ctx)
		},
	})
}
