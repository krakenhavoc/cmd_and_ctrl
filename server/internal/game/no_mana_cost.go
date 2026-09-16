package game

// no_mana_cost.go — CR 118.6: a spell with no mana cost can't be cast
// by paying it.
//
// "Some objects are described as having no mana cost. ... If an
// object has no mana cost, it can't be cast unless an alternative
// cost can be paid" (CR 118.6, 202.1b). Ancestral Vision, Living End,
// Hypergenesis and Lotus Bloom print nothing in the corner. ParseCost
// reads the empty string as the zero cost, which lands need, and
// which made every such card castable from hand for free.
//
// "No mana cost" is not "{0}". Ornithopter and Memnite print {0} and
// cast for nothing; the importer stamps that as the string "{0}", so
// the two stay distinct on Card.ManaCost.
//
// CR 118.6a keeps the exits open: an alternative cost may still be
// paid, and "cast it without paying its mana cost" is one. In this
// engine that is a claimed AlternativeCost, or an exile grant that
// carries its own price (cascade's and a Siege's "{0}", airbend's
// {2}). Both replace the printed cost before anything reads it.

// HasNoManaCost reports whether a card, as a spell, has no mana cost
// to pay (CR 118.6).
//
// Three conditions:
//
//   - The card is not a land. Lands are played, not cast (CR 305.1),
//     and an empty cost is how every land is printed.
//   - Its ManaCost, for the face the caller has materialised, is
//     empty. A modal DFC's land back is a land; a transform card's
//     back face has no cost (CR 712.8) and is reached only through a
//     grant that names its own price.
//   - Its Layout is set, meaning the card's printed fields were
//     stamped from a Scryfall record at deck import. Scryfall gives
//     every card a layout. A card with no Layout (a hand-built test
//     fixture, the demo seed) was never stamped, so an empty ManaCost
//     there means "unknown", not "no mana cost". The same posture as
//     Card.Colors, where empty means "not stamped" rather than
//     colourless. Those cards keep casting as they did.
func HasNoManaCost(card Card) bool {
	return card.Layout != "" && card.ManaCost == "" && !card.IsLand()
}

// castPaysPrintedCost reports whether this cast would pay the card's
// printed mana cost, which is the only way CR 118.6 can bite. A
// claimed alternative cost, or a live exile grant that names its own
// price, replaces the printed cost in printedCostLocked, so neither
// is a cast "by paying its mana cost".
func castPaysPrintedCost(alt *AlternativeCost, exileGrant ExilePlayPermission, hasExileGrant bool) bool {
	if alt != nil {
		return false
	}
	if hasExileGrant && exileGrant.CostOverride != "" {
		return false
	}
	return true
}
