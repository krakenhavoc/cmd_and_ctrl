package effects

// Bog Wreckage — Land (Odyssey, #5613):
//
//	"This land enters tapped.
//	 {T}: Add {B}.
//	 {T}, Sacrifice this land: Add one mana of any color."
//
// One of Odyssey's five sacrifice lands, built by odySacrificeLand:
// an enters-tapped replacement and two mana abilities, the second
// paying with the land itself.
//
// No simplification.
func init() {
	Register(odySacrificeLand("bce32541-3bc4-4553-ace2-784a58f164aa", "Bog Wreckage", "{B}"))
}
