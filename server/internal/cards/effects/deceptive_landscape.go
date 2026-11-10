package effects

// Deceptive Landscape — Land (Tarkir: Dragonstorm Commander):
//
//	"{T}: Add {C}.
//	 {T}, Sacrifice this land: Search your library for a basic Plains,
//	 Swamp, or Forest card, put it onto the battlefield tapped, then
//	 shuffle.
//	 Cycling {W}{B}{G}"
//
// The Landscape cycle's shape (rfReprintALandscape) for the Abzan colors.
//
// No simplification.
func init() {
	Register(rfReprintALandscape("1831fe12-dbe0-437f-8fc8-f01bbb701fe1", "Deceptive Landscape",
		"{W}{B}{G}", "Plains", "Swamp", "Forest"))
}
