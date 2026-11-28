package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// s82_old_b736_helpers.go — S82 OLD batch 736 (Portal Second Age, #5625).

// S82OldB736DamageThenGain is "<this> deals N damage to target opponent or
// planeswalker. You gain N life" (Kiss of Death): the damage first, then the
// life, which is gained whether or not the damage was prevented (CR 608.2c) but
// not at all when the only target is illegal on resolution (CR 608.2b).
func S82OldB736DamageThenGain(n int) func(*game.StackItem, *Context) error {
	return func(_ *game.StackItem, ctx *Context) error {
		target, ok := ctx.ClauseTarget(0)
		if !ok {
			return nil
		}
		if err := (DealDamage{Source: ctx.Source(), Target: target.ID, Amount: n}).Apply(ctx); err != nil {
			return err
		}
		return GainLife{Amount: n}.Apply(ctx)
	}
}

// S82OldB736DrawPerTappedCreature is "Draw a card for each tapped creature
// target opponent controls" (Theft of Dreams): counted as the spell resolves,
// and nothing when the only target has left (CR 608.2b).
func S82OldB736DrawPerTappedCreature(_ *game.StackItem, ctx *Context) error {
	target, ok := ctx.ClauseTarget(0)
	if !ok || target.Kind != game.TargetPlayer {
		return nil
	}
	n := S82OldB380TappedCreaturesOf(ctx.Game, target.ID)
	if n == 0 {
		return nil
	}
	return DrawCards{Player: ctx.Controller(), N: n}.Apply(ctx)
}
