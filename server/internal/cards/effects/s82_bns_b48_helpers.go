package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// s82B48DamageToTheFirstTarget is "<Spell> deals N damage to any target."
// as the first step of a spell that does more (Void Helix, Sarkhan's
// Rage): the damage goes to the first announced target.
func S82B48DamageToTheFirstTarget(amount int) func(ctx *Context) error {
	return func(ctx *Context) error {
		var target game.TargetRef
		if len(ctx.Item.Targets) > 0 {
			target = ctx.Item.Targets[0]
		}
		return DealDamage{Source: ctx.Source(), Target: target.ID, Amount: amount}.Apply(ctx)
	}
}
