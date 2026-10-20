package heuristic

// mill_cost.go — #3870: the price of a "Mill N cards" cost.
//
// #3845 gave activated abilities a mill component (Sinister
// Concoction's "Mill a card", Rot Farm Skeleton's "Mill four cards",
// Millikin's "{T}, Mill a card: Add {C}", which is not a mana ability,
// CR 605.1a). The params name no card, so
// until legal.MoveCost.Mill carried the count a policy read the library
// it paid as free.
//
// The bot already prices milling its own library, in one place: the
// dredge answer (dredge.go). A milled card leaves the library, where it
// was worth Weights.Library to the evaluation, and lands in the
// graveyard, where it is worth millValue: the mean fuel read of what is
// already there, so a graveyard of flashback and escape cards makes a
// mill worth more and a pile of dead lands keeps it at the floor. A mill
// cost is the same trade, so it is priced the same way, and the two
// answers cannot disagree about what a card off the top is worth.
//
// And the same cliff: a cost that would leave fewer than
// DredgeLibraryFloor cards in the library is declined, whatever the
// ability buys, because a short library is the CR 704.5b decking loss
// on a timer and the bot cannot read what the ability does well enough
// to trade its last draws for it.
//
// All of it sits behind Config.PriceMillCost; BaselineConfig reads the
// mill as free, as the heuristic did before.

// millCostDeclined is what a mill cost that would take the library
// under the floor is worth: below passing, so the policy never takes
// it, without the "would pay its last life" reason a CR 704.5a refusal
// carries. phyrexianLifeDeclined's shape.
const millCostDeclined = -1.0

// millCostShortLibrary reports whether paying a mill of n cards would
// leave the bot's library under DredgeLibraryFloor. False when the view
// does not carry the seat, which prices the mill and nothing more.
func (p *Policy) millCostShortLibrary(st *state, n int) bool {
	if !p.cfg.PriceMillCost || n <= 0 || st.seat == nil {
		return false
	}
	return st.seat.Library.Count-n < p.cfg.DredgeLibraryFloor
}

// millCostValue is what milling n cards off the bot's own library adds
// to a move: per card, the graveyard's fuel read less the library card
// it was. Positive when the graveyard wants the cards (it usually does:
// millValue's floor sits above Weights.Library), so a mill is a small
// reason to pay rather than a price; the floor above is the price.
func (p *Policy) millCostValue(st *state, n int) float64 {
	if !p.cfg.PriceMillCost || n <= 0 {
		return 0
	}
	return float64(n) * (p.millValue(st, nil) - st.w.Library)
}
