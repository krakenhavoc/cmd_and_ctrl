package effects

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_old_b702_helpers.go — shared shapes of S82 OLD batch 702 (Starter 1999,
// #5620).

// S82OldB702PathOfPeace is "Destroy target creature. Its owner gains 4 life."
// The owner is read before the destroy (the creature may be gone after it, CR
// 608.2h), and gains the life whether or not the destruction happened.
func S82OldB702PathOfPeace(_ *game.StackItem, ctx *Context) error {
	id, ok := B16FirstLegalTargetCard(ctx)
	if !ok {
		return nil
	}
	c, found := ctx.Game.LookupCardForEffect(id)
	if !found {
		return nil
	}
	if err := (DestroyTarget{Target: id}).Apply(ctx); err != nil {
		return err
	}
	return GainLife{Player: c.Owner, Amount: 4}.Apply(ctx)
}

// S82OldB702PsychicTransfer is "If the difference between your life total and
// target player's life total is 5 or less, exchange life totals with that
// player." The condition is judged as the spell resolves (CR 608.2c); a target
// that is no longer legal makes the spell do nothing (CR 608.2b).
func S82OldB702PsychicTransfer(_ *game.StackItem, ctx *Context) error {
	target, ok := FirstLegalPlayerTarget(ctx)
	if !ok {
		return nil
	}
	me, them := ctx.PlayerByID(ctx.Controller()), ctx.PlayerByID(target)
	if me == nil || them == nil {
		return nil
	}
	diff := me.Life - them.Life
	if diff < 0 {
		diff = -diff
	}
	if diff > 5 {
		return nil
	}
	return ExchangeLifeTotalsThen(ctx, me.ID, them.ID, game.KeyedTail{})
}

// S82OldB702Bargain is "Target opponent draws a card. You gain 7 life." The
// spell's only target gone by resolution means it does not resolve (CR
// 608.2b): neither half happens.
func S82OldB702Bargain(_ *game.StackItem, ctx *Context) error {
	t, ok := FirstLegalPlayerTarget(ctx)
	if !ok {
		return nil
	}
	if err := (DrawCards{Player: t, N: 1}).Apply(ctx); err != nil {
		return err
	}
	return GainLife{Player: ctx.Controller(), Amount: 7}.Apply(ctx)
}
