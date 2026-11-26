package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// DestroyTargetThenDamageItsController is "Destroy target <permanent>. <This
// spell> deals N damage to that permanent's controller" (Destructive
// Revelry, Peak Eruption). The controller is read before the destruction, so
// a permanent that dies (or survives, indestructible) still has its
// controller hit; a target that is illegal as the spell resolves stops the
// whole spell (CR 608.2b), which OnResolve is not run for.
func DestroyTargetThenDamageItsController(n int) func(item *game.StackItem, ctx *Context) error {
	return func(_ *game.StackItem, ctx *Context) error {
		id, ok := B16FirstLegalTargetCard(ctx)
		if !ok {
			return nil
		}
		victim, ok := ctx.Game.LookupCardForEffect(id)
		if !ok {
			return nil
		}
		controller := victim.Controller
		if err := (DestroyTarget{Target: id}).Apply(ctx); err != nil {
			return err
		}
		return DealDamage{Source: ctx.Source(), Target: controller, Amount: n}.Apply(ctx)
	}
}
