package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_old_b194_helpers.go — Cloudform and Lightform (Commander 2018, #5526):
//
//	"When this enchantment enters, it becomes an Aura with enchant creature.
//	 Manifest the top card of your library and attach this enchantment to it."

// BecomeAnAuraManifestAndAttach is that enters trigger, in printed order:
//
//  1. The enchantment becomes an Aura (CR 205.3, a layer-4 subtype change
//     from a resolved ability, CR 611.2c, CR 613.1d) pinned to this object
//     for as long as it is on the battlefield, and its enchant clause is
//     ADR 0154's swapped one, which names only the creature manifested here.
//     An Aura attached to nothing is put into its owner's graveyard (CR
//     704.5m), which is what happens with an empty library. If the
//     enchantment has already left, there is nothing to change.
//  2. The top card of the controller's library is manifested (CR 701.40a),
//     whether or not the enchantment is still there.
//  3. The enchantment is attached to the face-down creature, a new object
//     (CR 701.3c). An enchantment that has left is not.
type BecomeAnAuraManifestAndAttach struct{}

func (BecomeAnAuraManifestAndAttach) Apply(ctx *Context) error {
	g, item := ctx.Game, ctx.Item
	if item == nil {
		return nil
	}
	auraID := item.SourceCardID
	here := !g.AbilitySourceGoneForEffect(item)
	var aura game.Card
	if here {
		c, ok := g.LookupCardForEffect(auraID)
		if !ok {
			return nil
		}
		aura = c
		if err := (ScopedEffectFor{
			Target:   auraID,
			Mods:     []game.Mod{game.AddSubtypesMod("Aura")},
			Duration: g.PinnedTo(game.IndefiniteDuration(), auraID),
			Label:    aura.Name + " — it becomes an Aura",
		}).Apply(ctx); err != nil {
			return err
		}
		g.SwapEnchantClauseForEffect(auraID)
	}
	manifested, err := ManifestTopOfLibrary{}.Do(ctx)
	if err != nil || !here || manifested == uuid.Nil || !OnBattlefield(g, manifested) {
		return err
	}
	c, ok := g.LookupCardForEffect(manifested)
	if !ok {
		return nil
	}
	g.SetEnchantReturnedForEffect(auraID, game.ObjectRef{ID: manifested, Epoch: c.ObjectEpoch})
	return g.AttachForEffect(auraID, game.TargetRef{Kind: game.TargetCard, ID: manifested})
}
