package game

// cast_cost.go — which cost a cast is actually paying, and the one
// thing that settles on its own: CR 107.3b's value of X.
//
// A cast pays exactly one mana cost, and the engine already chooses
// it in one place (printedCostLocked): the alternative cost claimed
// at announce (CR 118.9), an exile grant's own price (cascade's
// "{0}", airbend's "{2}"), or the cost printed in the corner. Which
// one it is decides more than the number of mana owed, so the choice
// is a value the announce path can ask questions of rather than a
// string it parses twice.

// CastCost is the pair the announce path needs: what the card
// prints, and what this cast actually owes before the commander tax,
// the cost modifiers and convoke touch it.
//
// Both halves are cost STRINGS in Scryfall brace notation, because
// the question CR 107.3b asks is about the shape of the cost
// ("does it have an {X} slot") rather than about its total.
type CastCost struct {
	// Printed is the card's printed mana cost, for the face being
	// cast. Empty for a card with no mana cost (CR 202.1b).
	Printed string

	// Paid is the cost string this cast pays instead: the claimed
	// alternative cost, the exile grant's override, or Printed when
	// neither applies.
	Paid string
}

// CastCostFor picks the cost this cast pays, in the same order
// printedCostLocked applies it: the claimed alternative cost
// replaces the printed cost (CR 118.9), and a live exile grant's own
// price replaces whatever was chosen, because the grant belongs to
// the exiled INSTANCE and is the reason the cast is happening at all.
//
// Pure. `hasExileGrant` is the caller's "this cast is out of exile
// under a grant that names this player and has not expired" — the
// same flag castPaysPrintedCost takes, for the same reason.
func CastCostFor(card Card, altCostKey string, exileGrant ExilePlayPermission, hasExileGrant bool) CastCost {
	paid := alternativeCostString(card, altCostKey)
	if hasExileGrant && exileGrant.CostOverride != "" {
		paid = exileGrant.CostOverride
	}
	return CastCost{Printed: card.ManaCost, Paid: paid}
}

// LocksXAtZero reports whether CR 107.3b fixes X at 0 for this cast:
//
//	"If a player is casting a spell that has an {X} in its mana cost,
//	 the value of X isn't defined by the text of that spell, and an
//	 effect lets that player cast that spell while paying neither its
//	 mana cost nor an alternative cost that includes X, then the only
//	 legal choice for X is 0. This doesn't apply to effects that only
//	 reduce a cost, even if they reduce it to zero."
//
// Two reads of one pair of strings, and every free-cast path falls
// out of them:
//
//   - The printed cost has an {X} slot, so there is an X to announce
//     at all (CR 601.2b). A spell whose X lives somewhere OTHER than
//     the mana cost is untouched — Toxic Deluge's "pay X life"
//     additional cost and Waterbender's Restoration's waterbend {X}
//     are costs of their own, not the mana cost, so 107.3b has
//     nothing to say about them and their X is still announced on a
//     free cast.
//   - The cost being paid has no {X} slot. Cascade's "{0}" grant, a
//     Siege's free cast, "you may cast it without paying its mana
//     cost" as an AlternativeCost with an empty ManaCost, and an
//     alternative cost priced without X ("you may pay {R} rather
//     than pay this spell's mana cost") are all one answer here. An
//     alternative cost that DOES carry X ("{X}{R}, discard a card")
//     still asks, because the rule's exception is written about the
//     cost, not about the keyword granting it.
//
// Cost REDUCTIONS are outside this by construction: a modifier
// subtracts from the cost that was chosen here and never replaces
// it, so the {X} slot survives Ghalta's discount all the way down to
// {0} and the caster still announces X.
//
// CR 107.3c — a spell whose own text defines X — would be a third
// clause here, and there is deliberately none: no card in the
// catalog defines the X in its mana cost from its text, and the
// engine has no seam that could express it. If one arrives it
// attaches to the first read (the printed cost's X stops being an
// announcement), not to the second.
//
// A printed cost the parser cannot read (split and adventure cards
// import a joined "{1}{R} // {1}{U}") locks nothing. The engine
// refuses such a cast outright when it pays the printed cost, and
// when a grant prices it there is no {X} the parser can see to fix.
func (c CastCost) LocksXAtZero() bool {
	printed, err := ParseCost(c.Printed)
	if err != nil || printed.XSlots == 0 {
		return false
	}
	paid, err := ParseCost(c.Paid)
	if err != nil {
		return false
	}
	return paid.XSlots == 0
}
