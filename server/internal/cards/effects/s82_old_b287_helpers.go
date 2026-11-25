package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// OldB287PokesFoe is "this creature deals N damage to that creature" for a
// "blocks or becomes blocked by a creature" trigger (Kessig Forgemaster's
// Flameheart Werewolf, N = 2): the other creature of the block pair, hit by
// the source as an object, so one that has left still deals it.
func OldB287PokesFoe(n int) func(g *game.Game, item *game.StackItem) error {
	return func(g *game.Game, item *game.StackItem) error {
		if item.Trigger == nil {
			return nil
		}
		ctx := NewContext(g, item)
		ref, ok := ctx.SourceRef()
		foe := SkewerSlingerFoe(item.Trigger.Event, item.SourceCardID)
		if !ok || foe == uuid.Nil {
			return nil
		}
		return DealDamage{SourceObject: &ref, Target: foe, Amount: n}.Apply(ctx)
	}
}

// OldB287TargetDealsDamageToItself is "target creature deals damage to
// itself equal to its power" (Inner Struggle): the creature is the source of
// its own damage, read as the spell resolves (CR 608.2b, 120.3).
func OldB287TargetDealsDamageToItself(_ *game.StackItem, ctx *Context) error {
	id, ok := B16FirstLegalTargetCard(ctx)
	if !ok {
		return nil
	}
	c, ok := ctx.Game.LookupCardForEffect(id)
	if !ok || c.CurrentPower() <= 0 {
		return nil
	}
	return DealDamage{Source: id, Target: id, Amount: c.CurrentPower()}.Apply(ctx)
}

// OldB287DamageToYou reports whether damage to `target` goes to the
// source's controller, the player rather than their permanents (Goldnight Castigator).
func OldB287DamageToYou(target uuid.UUID, src *game.Card) bool {
	return target == src.Controller
}
