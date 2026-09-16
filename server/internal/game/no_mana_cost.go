package game

// no_mana_cost.go — CR 118.6: a spell with no mana cost can't be cast
// by paying it.
//
// Paraphrasing the rules, not quoting them: a card with no mana
// symbols where its mana cost would appear has no mana cost (CR
// 202.1b), and having no mana cost is an unpayable cost (CR 118.6).
// Attempting to cast such a spell is a legal action; attempting to
// pay the unpayable cost is not. Paper play would begin the cast and
// then reverse it at the payment step; this engine refuses at
// announce instead, which leaves the game in the same state.
// Ancestral Vision, Living End, Hypergenesis and Lotus Bloom print
// nothing in the corner. ParseCost reads the empty string as the zero
// cost, which lands need, and which made every such card castable
// from hand for free.
//
// "No mana cost" is not "{0}". Ornithopter and Memnite print {0} and
// cast for nothing; the importer stamps that as the string "{0}", so
// the two stay distinct on Card.ManaCost.
//
// CR 118.6a keeps the exits open: if an alternative cost is applied to
// the unpayable cost, including an effect that lets a player cast the
// spell without paying its mana cost, the alternative cost may be
// paid. A cost increase or an additional cost does not help; the cost
// stays unpayable. In this engine the alternative cost is a claimed
// AlternativeCost, or an exile grant that carries its own price
// (cascade's and a Siege's "{0}", airbend's {2}). Both replace the
// printed cost before anything reads it. Anything else (a tax, a
// kicker) is layered on the printed cost, so it is still refused.

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
