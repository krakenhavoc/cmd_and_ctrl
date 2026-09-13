package effects

// Horizon Canopy — Land (EDHREC rank 1692):
//
//	"{T}, Pay 1 life: Add {G} or {W}.
//	 {1}, {T}, Sacrifice this land: Draw a card."
//
// The original of the cycle Fiery Islet's file describes, and its
// Selesnya member: the life is a COST (validated before the land
// taps, refused at zero life), the two printed colours are not
// narrowed to the commander's identity, and the cash-in is Buried
// Ruin's three-component cost with a draw. Through the b15CanopyLand
// helper. See fiery_islet.go.
//
// No simplification.
func init() {
	Register(b15CanopyLand("262a5d83-506c-4781-9bc9-1a2b5d83955c", "Horizon Canopy", "G", "W"))
}
