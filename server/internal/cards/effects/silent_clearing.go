package effects

// Silent Clearing — Land (EDHREC rank 1653):
//
//	"{T}, Pay 1 life: Add {W} or {B}.
//	 {1}, {T}, Sacrifice this land: Draw a card."
//
// The Orzhov member of the Horizon Canopy cycle, Fiery Islet's shape
// with the colours swapped, through the b15CanopyLand helper: the
// life is a COST (validated before the land taps, refused at zero
// life), the two printed colours are not narrowed to the commander's
// identity, and the cash-in is Buried Ruin's three-component cost
// with a draw. See fiery_islet.go.
//
// No simplification.
func init() {
	Register(b15CanopyLand("fd45063f-c83c-431c-9104-f139c497ec0d", "Silent Clearing", "W", "B"))
}
