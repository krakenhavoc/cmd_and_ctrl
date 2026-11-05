package effects

// Shattered Landscape — Land (Star Trek Commander, #4974):
//
//	"{T}: Add {C}.
//	 {T}, Sacrifice this land: Search your library for a basic Mountain,
//	 Plains, or Swamp card, put it onto the battlefield tapped, then
//	 shuffle.
//	 Cycling {R}{W}{B}"
//
// Perilous Landscape's sibling for the Mardu colors; built by
// rfReprintALandscape.
//
// No simplification.
func init() {
	Register(rfReprintALandscape("7fbad3f2-66f6-4e3a-b9e0-1ddbcd94d42a", "Shattered Landscape",
		"{R}{W}{B}", "Mountain", "Plains", "Swamp"))
}
