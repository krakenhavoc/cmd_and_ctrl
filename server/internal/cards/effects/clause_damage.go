package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// clause_damage.go — "deals N damage to <one target> and M damage to
// <another target>" as one spell body (Arc Trail, Boulder Dash).

// damageEachClauseTarget is a spell's OnResolve that deals amounts[slot]
// damage from the spell to the pick of clause `slot`, in one damage
// instance. A slot whose target has left or stopped qualifying is skipped
// (CR 608.2b), so the other targets are still dealt their damage.
func damageEachClauseTarget(amounts ...int) func(*game.StackItem, *Context) error {
	return func(_ *game.StackItem, ctx *Context) error {
		return ctx.Game.DamageInstanceForEffect(func() error {
			for slot, n := range amounts {
				t, ok := ctx.ClauseTarget(slot)
				if !ok {
					continue
				}
				if err := (DealDamage{Source: ctx.Source(), Target: t.ID, Amount: n}).Apply(ctx); err != nil {
					return err
				}
			}
			return nil
		})
	}
}
