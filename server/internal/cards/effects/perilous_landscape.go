package effects

// Perilous Landscape — Land:
//
//	"{T}: Add {C}.
//	 {T}, Sacrifice this land: Search your library for a basic Island,
//	 Mountain, or Plains card, put it onto the battlefield tapped, then
//	 shuffle.
//	 Cycling {U}{R}{W}"
//
// A Panorama without the {1} that also cycles. The fetch is a
// searching ability, so the player picks the land (or declines), and
// the land enters tapped through the search's own entry clause.
//
// No simplification.
func init() {
	Register(rfReprintALandscape("e2b472dd-047d-47eb-9ebb-df6aa4b52dd4", "Perilous Landscape",
		"{U}{R}{W}", "Island", "Mountain", "Plains"))
}
