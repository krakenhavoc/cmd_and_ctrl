package effects

// Ravaged Highlands — Land (Odyssey, #5613):
//
//	"This land enters tapped.
//	 {T}: Add {R}.
//	 {T}, Sacrifice this land: Add one mana of any color."
//
// One of Odyssey's five sacrifice lands, built by odySacrificeLand:
// an enters-tapped replacement and two mana abilities, the second
// paying with the land itself.
//
// No simplification.
func init() {
	Register(odySacrificeLand("68919931-b995-4aa0-9d2a-e80eae2f335f", "Ravaged Highlands", "{R}"))
}
