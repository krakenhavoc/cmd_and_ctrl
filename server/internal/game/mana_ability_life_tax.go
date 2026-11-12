package game

// mana_ability_life_tax.go — "Mana abilities of this land cost an
// additional 1 life to activate" (#5661, Thran Portal).
//
// The clause is a cost increase on every mana ability of one permanent,
// and the abilities it prices are mostly not printed on that permanent:
// Thran Portal is a basic land type of its controller's choice "in
// addition to its other types", so the ability it taps with is the
// intrinsic one of that type (CR 305.6), which no card file declares
// and so no card file can price. A mana ability another effect grants
// the land (Chromatic Lantern) is an ability "of this land" too.
//
// CR 118.8: "An additional cost is a cost listed in a spell's rules
// text, or applied to a spell or ability from another effect, that its
// controller must pay at the same time they pay the spell's mana cost
// or the ability's activation cost." CR 602.2b sends an activation
// through CR 601.2b–i, and CR 601.2f totals "the mana cost or
// alternative cost ..., plus all additional costs and cost increases".
// So the life is part of the total cost of each activation: checked
// with the rest before anything is paid (CR 119.4, a player at 0 life
// cannot tap the land for mana) and paid with it.
//
// # Where it lives
//
// The surcharge is a number on the permanent's layered view,
// Characteristic.ManaAbilityLifeTax, written by the static ability's
// Apply the way a restriction is (restrictions.go): the layer pass is
// the one thing that knows which statics are working, so a land that
// lost its abilities (Blood Moon, CR 305.7) stops charging without a
// rule here. Like a restriction it is only ever added to and no layer
// clears it, so its layer and timestamp do not matter, with one
// exception the reader makes below.
//
// manaAbilityRows, the one accessor every reader of a permanent's mana
// abilities goes through, adds it to each row's LifeCost. From there on
// it IS a Mana Confluence-style life cost, and everything that already
// understands one does the right thing unchanged: the activation's
// validate-then-pay (activateManaAbilityLocked), the wire's life_cost
// chip, the auto-tapper's pain tier (#2392) and its CanPayLife gate,
// LandManaOptionsForEffect, and internal/legal's enumerator for a bot.
//
// The rows a borrow LENDS from this permanent to another are read from
// the catalog (borrowedManaRows), not from here, so the other permanent
// pays only its own surcharge: the text prices abilities "of this
// land".

// manaAbilityLifeTaxOf is the life every mana ability of `c` costs on
// top of its own cost right now. Zero off the battlefield (no layered
// view), and zero once the permanent's own abilities are gone: the
// surcharge is one of them, and a removal sorted after it in the same
// layer has to take it too (CR 613.1f), which a field no layer clears
// would not do by itself.
func manaAbilityLifeTaxOf(c *Card) int {
	if c.effective == nil || c.effective.AbilitiesRemoved {
		return 0
	}
	return c.effective.ManaAbilityLifeTax
}

// taxedManaRows is `rows` with `tax` more life on each. A copy: the
// input may be the catalog's own slice or a shared intrinsic row
// (intrinsicLandRows).
func taxedManaRows(rows []ManaAbilityShape, tax int) []ManaAbilityShape {
	if tax <= 0 || len(rows) == 0 {
		return rows
	}
	out := make([]ManaAbilityShape, len(rows))
	copy(out, rows)
	for i := range out {
		out[i].LifeCost += tax
	}
	return out
}
