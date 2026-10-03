package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// enchanted_planeswalker.go — the shared vocabulary of an Aura that
// enchants a planeswalker (ADR 0109 §2, owner decision 2). The five
// Talents print the same three shapes: a loyalty ability granted to the
// enchanted planeswalker (GrantAbilitiesToAttached over a bundle whose
// row has a LoyaltyCost — nothing else to declare, since CR 606.3 and
// 606.6 are counted on the permanent that has the row), a trigger that
// watches that planeswalker, and an effect that puts loyalty on it.
//
// "Enchanted planeswalker" is read off the Aura each time: as a trigger
// is harvested (its AttachedTo then), and at resolution through
// SourcePermanent, so an Aura destroyed in response still names the
// planeswalker it was on (CR 608.2h).

// YouActivatedALoyaltyAbilityOfEnchanted — "whenever you activate a
// loyalty ability of enchanted planeswalker" (Elspeth's Talent, Rowan's
// Talent). "You" is the Aura's controller; the ability is a loyalty
// ability of the permanent the Aura is attached to, printed or granted,
// which is the activation event's Loyalty bit (CR 606.2).
//
// A copy is not activated (CR 707.10: "a copy of an activated ability
// isn't activated") and announces no activation, so a copied loyalty
// ability never triggers this — which is what keeps Rowan's Talent from
// copying its own copy.
func YouActivatedALoyaltyAbilityOfEnchanted(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) bool {
	return ev.Kind == game.EventActivateAbility && ev.Loyalty &&
		ev.Actor == source.Controller &&
		source.AttachedTo.Kind == game.TargetCard && ev.CardID == source.AttachedTo.ID
}

// WheneverYouActivateALoyaltyAbilityOfEnchanted is the printed shape.
// The trigger goes on the stack above the loyalty ability (CR 603.3b)
// and resolves first; the ability is named by the triggering event's
// StackItemID.
func WheneverYouActivateALoyaltyAbilityOfEnchanted(label string, effect Effect) game.TriggeredAbility {
	return On(game.EventActivateAbility, YouActivatedALoyaltyAbilityOfEnchanted, label, effect)
}

// ACreatureDealtDamageToEnchanted — "whenever a creature deals damage
// to enchanted planeswalker" (Liliana's Talent). One event per source,
// so two creatures are two triggers. The source has to be a creature
// on the battlefield as the damage is dealt, as No Mercy reads it.
func ACreatureDealtDamageToEnchanted(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
	if ev.Kind != game.EventDealDamage || ev.Amount <= 0 ||
		source.AttachedTo.Kind != game.TargetCard || ev.Target != source.AttachedTo.ID {
		return false
	}
	return damageSourceIsABattlefieldCreature(ev, g)
}

// putLoyaltyCounterOnEnchantedPlaneswalker is "put a loyalty counter on
// enchanted planeswalker" (Teferi's Talent, Vivien's Talent). The
// counter lands only while that planeswalker is still on the
// battlefield.
func putLoyaltyCounterOnEnchantedPlaneswalker(g *game.Game, item *game.StackItem) error {
	aura, ok := NewContext(g, item).SourcePermanent()
	if !ok || aura.AttachedTo.Kind != game.TargetCard {
		return nil
	}
	host := aura.AttachedTo.ID
	if z := g.FindCardZoneForEffect(host); z == nil || z.Kind != game.ZoneBattlefield {
		return nil
	}
	return g.AddCounterForEffect(host, game.CounterLoyalty, 1)
}
