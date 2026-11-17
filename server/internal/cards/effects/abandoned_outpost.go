package effects

// Abandoned Outpost — Land (Odyssey, #5613):
//
//	"This land enters tapped.
//	 {T}: Add {W}.
//	 {T}, Sacrifice this land: Add one mana of any color."
//
// One of Odyssey's five sacrifice lands, built by odySacrificeLand:
// an enters-tapped replacement and two mana abilities, the second
// paying with the land itself.
//
// No simplification.
func init() {
	Register(odySacrificeLand("5a70ccfa-d12d-4e62-a1a4-f05cda2fd442", "Abandoned Outpost", "{W}"))
}
