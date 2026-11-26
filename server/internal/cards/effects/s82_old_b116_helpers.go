package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// s82_old_b116_helpers.go — shared pieces of Ikoria: Lair of Behemoths
// batch 116 (#5512).

// S82OldB116RevealACreatureCardCost is "As an additional cost to cast this
// spell, reveal a creature card from your hand" where the spell then reads the
// revealed card's power (Charge of the Forever-Beast): a reveal of a creature
// card from the hand only (no permanent may be chosen instead, unlike
// ChooseOrRevealCost's behold), whose pick is recorded on the cast's payment
// (RevealCost.Chosen) for B25ChosenPower.
func S82OldB116RevealACreatureCardCost() *game.AdditionalCost {
	return &game.AdditionalCost{
		Reveal: &game.RevealCost{CardType: "creature", Chosen: true},
		Label:  "Reveal a creature card from your hand",
	}
}

// S82OldB116DamageToTargetCreatureExcessToItsController is "<source> deals
// `amount` damage to target creature. Excess damage is dealt to that
// creature's controller instead" (Flame Spill): the split is made before the
// damage is dealt (CR 120.4a), what the creature still needs to die (toughness
// less damage already marked) going to it and the rest to its controller, as
// one damage event. A creature whose toughness can't be read takes all of it.
func S82OldB116DamageToTargetCreatureExcessToItsController(item *game.StackItem, ctx *Context, amount int) error {
	t, ok := ctx.ClauseTarget(0)
	if !ok || t.Kind != game.TargetCard {
		return nil
	}
	c, ok := ctx.Game.LookupCardForEffect(t.ID)
	if !ok {
		return nil
	}
	toCreature, toController := amount, 0
	if c.ToughnessIsKnown() {
		need := max(c.CurrentToughness()-c.DamageMarked, 0)
		toCreature = min(amount, need)
		toController = amount - toCreature
	}
	return ctx.Game.DamageInstanceForEffect(func() error {
		if toCreature > 0 {
			if err := (DealDamage{Source: item.SourceCardID, Target: t.ID, Amount: toCreature}).Apply(ctx); err != nil {
				return err
			}
		}
		if toController > 0 {
			return ctx.Game.DealDamageToPlayerForEffect(item.SourceCardID, c.Controller, toController)
		}
		return nil
	})
}

// S82OldB116DamageEqualToTheChosenPower is "<source> deals damage to target
// <thing> equal to the revealed card's power" (Charge of the Forever-Beast):
// the first legal target takes the power the cast's reveal recorded.
func S82OldB116DamageEqualToTheChosenPower(_ *game.StackItem, ctx *Context) error {
	target, ok := B16FirstLegalTargetCard(ctx)
	if !ok {
		return nil
	}
	n, ok := B25ChosenPower(ctx)
	if !ok {
		return nil
	}
	return DealDamage{Source: ctx.Source(), Target: target, Amount: n}.Apply(ctx)
}
