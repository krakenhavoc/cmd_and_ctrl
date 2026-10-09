package effects

// Contaminated Landscape — Land:
//
//	"{T}: Add {C}.
//	 {T}, Sacrifice this land: Search your library for a basic Plains,
//	 Island, or Swamp card, put it onto the battlefield tapped, then
//	 shuffle.
//	 Cycling {W}{U}{B}"
//
// Perilous Landscape's sibling for the Esper colors; both are built by
// rfReprintALandscape.
//
// No simplification.
func init() {
	Register(rfReprintALandscape("28196fd9-00c9-4cd0-b603-0eec8511ec79", "Contaminated Landscape",
		"{W}{U}{B}", "Plains", "Island", "Swamp"))
}
