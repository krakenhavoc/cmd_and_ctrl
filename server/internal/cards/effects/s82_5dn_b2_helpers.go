package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_5dn_b2_helpers.go — the shape Fifth Dawn batch 2 shares (#5604).

// dn5B2BoostTheTargetThenScry2 is "Target creature gets +P/+T until end of
// turn. Scry 2." (Ferocious Charge, Stand Firm, Lose Hope): Titan's
// Strength's order, the boost first and the scry last, so the scry's prompt
// has nothing printed after it to hold. A spell that resolves always scries;
// one whose only target became illegal doesn't resolve (CR 608.2b).
func Dn5B2BoostTheTargetThenScry2(power, toughness int, label string) func(item *game.StackItem, ctx *Context) error {
	return func(_ *game.StackItem, ctx *Context) error {
		if id, ok := B16FirstLegalTargetCard(ctx); ok {
			if err := (BoostUntilEOT{Target: id, Power: power, Toughness: toughness, Label: label}).Apply(ctx); err != nil {
				return err
			}
		}
		return Scry{Player: ctx.Controller(), N: 2}.Apply(ctx)
	}
}

// GraftedSacrificeOnUnattach is "Whenever this Equipment becomes unattached
// from a permanent, sacrifice that permanent.", shared by Grafted
// Exoskeleton and Grafted Wargear, which print the same clause. The label is
// the row's Key.
func GraftedSacrificeOnUnattach(label string) game.TriggeredAbility {
	return OnAny([]game.EventKind{game.EventUnattach, game.EventLTB}, Self, label,
		Do(Instruction(func(ctx *Context) error {
			host := graftedHost(ctx)
			if host == uuid.Nil {
				return nil
			}
			if z := ctx.Game.FindCardZoneForEffect(host); z == nil || z.Kind != game.ZoneBattlefield {
				return nil
			}
			return SacrificePermanent{Target: host}.Apply(ctx)
		})))
}

// graftedHost is the permanent the Equipment became unattached from: the
// event's target when the link broke with the Equipment still on the
// battlefield, and what the departed Equipment was last attached to when it
// left.
func graftedHost(ctx *Context) uuid.UUID {
	tc := ctx.Trigger()
	if tc.Event.Kind == game.EventUnattach {
		return tc.Event.Target
	}
	if info, ok := ctx.TriggeringPermanent(); ok && info.AttachedTo.Kind == game.TargetCard {
		return info.AttachedTo.ID
	}
	return uuid.Nil
}

// VendettaResolve is "Destroy target … creature. It can't be regenerated.
// You lose life equal to that creature's toughness.", shared by Vendetta and
// Devour in Shadow, which prints it for any creature.
func VendettaResolve(_ *game.StackItem, ctx *Context) error {
	for _, t := range ctx.LegalTargets() {
		if t.Kind != game.TargetCard {
			continue
		}
		c, ok := ctx.Game.LookupCardForEffect(t.ID)
		if !ok {
			return nil
		}
		toughness := c.CurrentToughness()
		if err := (DestroyTarget{Target: t.ID, CantBeRegenerated: true}).Apply(ctx); err != nil {
			return err
		}
		if toughness <= 0 {
			return nil
		}
		return ctx.Game.ChangePlayerLifeForEffect(ctx.Source(), ctx.Controller(), -toughness)
	}
	return nil
}
