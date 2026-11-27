package effects

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_old_b648_helpers.go — Invasion, "Kicker" (S82 OLD batch 648, #5616).
// Append-only.

// S82OldB648OverloadResolve is Overload: "Destroy target artifact if its mana
// value is 2 or less. If this spell was kicked, destroy that artifact if its
// mana value is 5 or less instead." The artifact is any artifact when it is
// targeted; the limit is checked as the spell resolves, on the target still
// legal then (CR 608.2b), so a bigger artifact is a legal target that is
// simply not destroyed.
func S82OldB648OverloadResolve(_ *game.StackItem, ctx *Context) error {
	id, ok := B16FirstLegalTargetCard(ctx)
	if !ok {
		return nil
	}
	card, found := ctx.Game.LookupCardForEffect(id)
	if !found {
		return nil
	}
	limit := 2
	if ctx.WasKicked() {
		limit = 5
	}
	if card.ManaValue() > limit {
		return nil
	}
	return DestroyTarget{Target: id}.Apply(ctx)
}

// S82OldB648AgonizingDemiseResolve is "Destroy target nonblack creature. It
// can't be regenerated. If this spell was kicked, Agonizing Demise deals
// damage equal to that creature's power to the creature's controller." The
// power and the controller are read before the destroy, so they are the
// creature's last known information whether or not it died (CR 608.2h);
// an indestructible creature still costs its controller the damage. A power of
// zero or less deals no damage (DealDamage does nothing for it, CR 120.8).
func S82OldB648AgonizingDemiseResolve(_ *game.StackItem, ctx *Context) error {
	id, ok := B16FirstLegalTargetCard(ctx)
	if !ok {
		return nil
	}
	card, found := ctx.Game.LookupCardForEffect(id)
	if !found {
		return nil
	}
	power, controller := card.CurrentPower(), card.Controller
	if err := (DestroyTarget{Target: id, CantBeRegenerated: true}).Apply(ctx); err != nil {
		return err
	}
	if !ctx.WasKicked() {
		return nil
	}
	return DealDamage{Source: ctx.Source(), Target: controller, Amount: power}.Apply(ctx)
}

// S82OldB648HypnoticCloudResolve is "Target player discards a card. If this
// spell was kicked, that player discards three cards instead."
func S82OldB648HypnoticCloudResolve(item *game.StackItem, ctx *Context) error {
	n := 1
	if ctx.WasKicked() {
		n = 3
	}
	return TargetPlayerDiscards(n)(ctx.Game, item)
}

// S82OldB648DestroyTheTargetNoRegen is the enters trigger "destroy target
// nonblack creature. It can't be regenerated" (Shivan Emissary).
func S82OldB648DestroyTheTargetNoRegen(ctx *Context) error {
	return DestroyTheTargetPermanentNoRegen(ctx.Item, ctx)
}

// S82OldB648DestroyTheTarget is the enters trigger "destroy target
// enchantment" (Tolarian Emissary).
func S82OldB648DestroyTheTarget(ctx *Context) error {
	return DestroyTheTargetPermanent(ctx.Item, ctx)
}
