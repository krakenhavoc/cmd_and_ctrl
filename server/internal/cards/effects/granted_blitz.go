package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// granted_blitz.go — #5667: blitz a permanent gives the spells its
// controller casts, and a reduction of blitz costs alone. Henzie
// "Toolbox" Torre: "Each creature spell you cast with mana value 4 or
// greater has blitz. The blitz cost is equal to its mana cost." and
// "Blitz costs you pay cost {1} less for each time you've cast your
// commander from the command zone this game."
//
// CR 702.152a: "Blitz represents three abilities: two static abilities
// that function while the card with blitz is on the stack, one of which
// may create a delayed triggered ability, and a static ability that
// functions while the object with blitz is on the battlefield. "Blitz
// [cost]" means "You may cast this card by paying [cost] rather than its
// mana cost," "If this spell's blitz cost was paid, sacrifice the
// permanent this spell becomes at the beginning of the next end step,"
// and "As long as this permanent's blitz cost was paid, it has haste and
// 'When this permanent is put into a graveyard from the battlefield, draw
// a card.'""
//
// The first is a granted alternative cost (ADR 0118 §3,
// granted_alternative_cost.go). The other two hang on "this spell's blitz
// cost was paid", not on the permanent that gave the spell blitz, so the
// engine reads them off the claim's key (game.GrantedAltCostBlitz) as the
// permanent enters, the way granted evoke's sacrifice is read: a creature
// blitzed through Henzie has haste, draws when it dies and is sacrificed
// at the next end step even if Henzie left while the spell was on the
// stack.

// GrantedAltCostBlitz is the key of a granted blitz offer. It is
// game.GrantedAltCostBlitz: the entry hook reads it. An on-disk identity.
const GrantedAltCostBlitz = game.GrantedAltCostBlitz

// BlitzForCreatureSpellsYouCast is "Each creature spell you cast with
// mana value `minManaValue` or greater has blitz. The blitz cost is equal
// to its mana cost." (Henzie "Toolbox" Torre): a granted alternative cost
// priced at the spell's own mana cost as it is cast
// (GrantedAlternativeCost.PricedAtManaCost), which carries blitz's haste,
// draw and end-step sacrifice (CR 702.152a). It reaches every zone the
// printed mana cost could be paid from (CR 118.9a), a commander in the
// command zone included, which still pays its tax on top (CR 118.9d). A
// creature that prints its own blitz keeps it as a second offer, and the
// caster picks one (CR 702.152b). A creature with {X} in its mana cost
// keeps the {X} in its blitz cost and X is chosen as usual (CR 107.3a);
// the mana value the floor is judged by is the card's before X is
// announced, so that {X} counts as 0.
func BlitzForCreatureSpellsYouCast(minManaValue int) game.GrantedAlternativeCost {
	return game.GrantedAlternativeCost{
		Offer: game.AlternativeCost{
			Key:   GrantedAltCostBlitz,
			Label: "Blitz",
			Blitz: true,
		},
		Spells:           game.PermissionFilter{CreatureOnly: true},
		MinManaValue:     &minManaValue,
		PricedAtManaCost: true,
	}
}

// ForItsBlitzCost passes on a spell cast for a blitz cost, the card's
// own or one a permanent gave it (game.CostQuery.AltBlitz): the "blitz
// costs" of "Blitz costs you pay cost {1} less". The same spell cast for
// its mana cost, or for any other alternative cost, does not pass.
func ForItsBlitzCost() CostPredicate {
	return func(q game.CostQuery) bool { return q.AltBlitz }
}

// BlitzCostsYouPayCostLessForEachCommanderCast is "Blitz costs you pay
// cost {1} less for each time you've cast your commander from the command
// zone this game." (Henzie "Toolbox" Torre): a reduction of generic mana
// (CR 118.7a) that applies to a blitz cost its controller pays, printed
// or granted, and to nothing else. The count is the controller's
// commander-cast tally as the cost is totalled (CR 601.2f), so a
// commander being cast from the command zone for its blitz cost is not
// yet counted for itself.
func BlitzCostsYouPayCostLessForEachCommanderCast() game.CostModifier {
	return CostsLessEach(func(q game.CostQuery) int {
		return socB3CommanderCastsFromCommandZone(q.Game, q.Source.Controller)
	}, "Blitz costs you pay cost {1} less for each time you've cast your commander from the command zone this game.",
		YourSpell(), ForItsBlitzCost())
}
