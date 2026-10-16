package heuristic

// library_exile_cost.go — #3966: the price of an "Exile the top N cards
// of your library" cost.
//
// ADR 0109 §7 gave activated abilities that component (Seasoned
// Tactician's four, Arc-Slogger's ten, Phyrexian Devourer's one). The
// params name no card, so until legal.MoveCost.LibraryExile carried the
// count a policy read the library it paid as free.
//
// It is a mill (mill_cost.go) without the graveyard fuel: the cards leave
// the library, where each was worth Weights.Library to the evaluation,
// and go to exile, where the bot cannot cast or return them. So each card
// costs its Library weight and buys nothing back. The price that matters
// is the mill's floor, held the same: a cost that would leave fewer than
// DredgeLibraryFloor cards in the library is declined, whatever the
// ability buys, because a short library is the CR 704.5b decking loss on
// a timer.
//
// All of it sits behind Config.PriceLibraryExileCost; BaselineConfig
// reads the exile as free, as the heuristic did before.

// libraryExileCostShortLibrary reports whether paying an exile of the
// top n cards would leave the bot's library under DredgeLibraryFloor.
// False when the view does not carry the seat, which prices the cards
// and nothing more.
func (p *Policy) libraryExileCostShortLibrary(st *state, n int) bool {
	if !p.cfg.PriceLibraryExileCost || n <= 0 || st.seat == nil {
		return false
	}
	return st.seat.Library.Count-n < p.cfg.DredgeLibraryFloor
}

// libraryExileCostValue is what exiling the top n cards of the bot's own
// library costs a move: the library card each was, with no graveyard to
// land in. costValue subtracts it.
func (p *Policy) libraryExileCostValue(st *state, n int) float64 {
	if !p.cfg.PriceLibraryExileCost || n <= 0 {
		return 0
	}
	return float64(n) * st.w.Library
}
