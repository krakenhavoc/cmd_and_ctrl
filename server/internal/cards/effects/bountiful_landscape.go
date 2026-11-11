package effects

// Bountiful Landscape — Land (Tarkir: Dragonstorm Commander):
//
//	"{T}: Add {C}.
//	 {T}, Sacrifice this land: Search your library for a basic Forest,
//	 Island, or Mountain card, put it onto the battlefield tapped, then
//	 shuffle.
//	 Cycling {G}{U}{R}"
//
// The Landscape cycle's shape (rfReprintALandscape) for the Temur colors.
//
// No simplification.
func init() {
	Register(rfReprintALandscape("2eb69a8f-9456-4852-b868-85ae609d3441", "Bountiful Landscape",
		"{G}{U}{R}", "Forest", "Island", "Mountain"))
}
