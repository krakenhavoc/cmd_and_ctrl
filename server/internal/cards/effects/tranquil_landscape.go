package effects

// Tranquil Landscape — Land (Star Trek Commander, #4974):
//
//	"{T}: Add {C}.
//	 {T}, Sacrifice this land: Search your library for a basic Forest,
//	 Plains, or Island card, put it onto the battlefield tapped, then
//	 shuffle.
//	 Cycling {G}{W}{U}"
//
// Perilous Landscape's sibling for the Bant colors; built by
// rfReprintALandscape.
//
// No simplification.
func init() {
	Register(rfReprintALandscape("d4eb65d5-99fd-4daf-b7d3-8ebf99ee9c61", "Tranquil Landscape",
		"{G}{W}{U}", "Forest", "Plains", "Island"))
}
