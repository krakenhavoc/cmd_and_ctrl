package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// granted_miracle.go — #4685: the statics that give the cards in their
// controller's hand miracle (CR 702.94a) at a price, narrowed by a card
// filter. The engine half is game/miracle.go: the reveal trigger is grown
// from the offer (buildDef), fires only for the cards THIS static reaches,
// and carries the price onto the cast permission its resolution grants.
//
// The shape is the general one a granted keyword with a cost needs — a
// GrantedAlternativeCost limited by Zones to the zone the cards sit in and
// by Spells to the cards it reaches — for a zone the cards may already be
// cast from. #4680's granted sneak for creature cards in a graveyard also
// has to open the zone, so it is a standing permission instead
// (CreatureCardsInYourGraveyardHaveSneak, ADR 0025 amendment 2026-11-04).

// grantedMiracleInYourHand is "<cards> in your hand have miracle <cost>":
// the granted miracle offer at `cost`, claimable only under the hand
// permission a revealed trigger's resolution grants (RequiresGrant), over
// the cards `spells` reaches. `text` is the printed subject for the
// ability row ("Each instant and sorcery card in your hand"); empty is
// Molecule Man's "Nonland cards in your hand".
func grantedMiracleInYourHand(cost string, spells game.PermissionFilter, text string) game.GrantedAlternativeCost {
	return game.GrantedAlternativeCost{
		Offer: game.AlternativeCost{
			Key:           GrantedAltCostMiracle,
			Label:         "Miracle " + cost,
			ManaCost:      cost,
			RequiresGrant: true,
		},
		Zones:      []game.ZoneKind{game.ZoneHand},
		Spells:     spells,
		SpellsText: text,
	}
}

// MiracleForInstantAndSorceryCardsInYourHand is Lorehold, the Historian's
// static: "Each instant and sorcery card in your hand has miracle {cost}."
// A creature, artifact or land card is never offered the reveal.
func MiracleForInstantAndSorceryCardsInYourHand(cost string) game.GrantedAlternativeCost {
	return grantedMiracleInYourHand(cost, game.PermissionFilter{InstantOrSorceryOnly: true},
		"Each instant and sorcery card in your hand")
}

// MiracleForEnchantmentCardsInYourHandReducedBy is Aminatou, Veil Piercer's
// static: "Each enchantment card in your hand has miracle. Its miracle cost
// is equal to its mana cost reduced by {N}." The price is the card's own
// mana cost (PricedAtManaCost) less `n` generic mana, so a coloured symbol
// is never reduced and the price floors at {0} (CR 118.7).
func MiracleForEnchantmentCardsInYourHandReducedBy(n int) game.GrantedAlternativeCost {
	g := grantedMiracleInYourHand("", game.PermissionFilter{CardType: "Enchantment"},
		"Each enchantment card in your hand")
	g.PricedAtManaCost = true
	g.ManaCostReducedBy = n
	return g
}
