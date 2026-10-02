package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// exile_if_dies_g1.go — shared bodies for ADR 0108 PR 1, group 1 (#1886,
// #1887), the ones exile_if_dies.go does not already have.

// shrinkFirstTargetExileIfItDies is "Target creature gets -N/-N until end
// of turn. If that creature would die this turn, exile it instead."
// (Bleed Dry, Ob Nixilis's Cruelty). Printed order is resolution order:
// the shrink first, then the replacement, so a creature the shrink kills
// is exiled.
func shrinkFirstTargetExileIfItDies(n int) func(item *game.StackItem, ctx *Context) error {
	return func(_ *game.StackItem, ctx *Context) error {
		id, ok := b16FirstLegalTargetCard(ctx)
		if !ok {
			return nil
		}
		if err := (BoostUntilEOT{Target: id, Power: -n, Toughness: -n}).Apply(ctx); err != nil {
			return err
		}
		return ExileIfItWouldDieThisTurn{Target: id}.Apply(ctx)
	}
}

// damageFirstTargetThenNoRegen is "~ deals N damage to target creature.
// It can't be regenerated this turn." (Engulfing Flames, Rage of
// Purphoros). The rider is the spell's, not the damage's, so it holds
// even when the damage is prevented (ADR 0108 §2). `then` runs last, for
// a card with a further clause (Rage of Purphoros's scry).
func damageFirstTargetThenNoRegen(amount int, then func(ctx *Context) error) func(item *game.StackItem, ctx *Context) error {
	return func(_ *game.StackItem, ctx *Context) error {
		id, ok := b16FirstLegalTargetCard(ctx)
		if !ok {
			return nil
		}
		if err := (DealDamage{Source: ctx.Source(), Target: id, Amount: amount}).Apply(ctx); err != nil {
			return err
		}
		if err := (CantBeRegeneratedThisTurn{Target: id}).Apply(ctx); err != nil {
			return err
		}
		if then == nil {
			return nil
		}
		return then(ctx)
	}
}

// targetCantBeRegeneratedThisTurn is "Target creature can't be
// regenerated this turn." as a whole effect (Hurr Jackal, Furnace Brood,
// Gravebind).
func targetCantBeRegeneratedThisTurn(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	id, ok := b16FirstLegalTargetCard(ctx)
	if !ok {
		return nil
	}
	return CantBeRegeneratedThisTurn{Target: id}.Apply(ctx)
}
