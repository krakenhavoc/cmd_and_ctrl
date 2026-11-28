package effects

import (
	"errors"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_old_b774_helpers.go — shared pieces of S82 OLD batch 774 (Weatherlight
// Auras, #5629).

// OldB774ReturnTheEventCardUnderItsOwnersControl is "return that card to the
// battlefield under its owner's control" for a trigger whose event names the
// card that left the battlefield (Abduction): the card returns as a new object
// (CR 400.7) under its OWNER, not under the trigger's controller, and a card
// that has since left the graveyard is nothing to return.
func OldB774ReturnTheEventCardUnderItsOwnersControl(g *game.Game, item *game.StackItem) error {
	if item.Trigger == nil {
		return nil
	}
	id := item.Trigger.Event.CardID
	card, ok := g.LookupCardForEffect(id)
	if !ok {
		return nil
	}
	if z := g.FindCardZoneForEffect(id); z == nil || z.Kind != game.ZoneGraveyard {
		return nil
	}
	_, err := g.ReturnToBattlefieldForEffect(id, card.Owner, false)
	if errors.Is(err, game.ErrCardNotFound) {
		return nil
	}
	return err
}

// OldB774MayDiscardAtRandomToUntapEnchantedCreature is "that player may discard
// a card at random. If the player does, untap that creature." (Apathy), the
// upkeep trigger of the enchanted creature's controller. The player is the
// upkeep's, who pays; the creature is the one the Aura enchants as the trigger
// resolves, untapped only while it is still that object (CR 400.7).
func OldB774MayDiscardAtRandomToUntapEnchantedCreature(name string) Effect {
	return func(g *game.Game, item *game.StackItem) error {
		if item == nil || item.Trigger == nil {
			return nil
		}
		ctx := NewContext(g, item)
		aura, ok := ctx.SourcePermanent()
		if !ok || aura.AttachedTo.Kind != game.TargetCard {
			return nil
		}
		host, ok := g.LookupCardForEffect(aura.AttachedTo.ID)
		if !ok || !OnBattlefield(g, host.InstanceID) {
			return nil
		}
		return MayPay{
			Chooser:      item.Trigger.Event.Actor,
			Pay:          RandomDiscardPayment(1),
			Question:     name + " — discard a card at random to untap the enchanted creature?",
			Then:         MayPayUntapEnchantedCreaturePaid,
			CarryObjects: []game.ObjectRef{{ID: host.InstanceID, Epoch: host.ObjectEpoch}},
		}.Apply(ctx)
	}
}

// OldB774EnchantedCreatureGetsUntilEOT is "Enchanted creature gets +P/+T until
// end of turn" in an Aura's own ability (Briar Shield, Nature's Kiss): the
// creature the Aura is attached to as the ability resolves, or was attached to
// when it left (CR 608.2h), still on the battlefield.
func OldB774EnchantedCreatureGetsUntilEOT(power, toughness int, label string) Effect {
	return func(g *game.Game, item *game.StackItem) error {
		ctx := NewContext(g, item)
		host, ok := EnchantedCreatureAsItResolves(ctx)
		if !ok || !OnBattlefield(g, host) {
			return nil
		}
		return BoostUntilEOT{Target: host, Power: power, Toughness: toughness, Label: label}.Apply(ctx)
	}
}
