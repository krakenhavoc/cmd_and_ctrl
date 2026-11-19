package effects

// Twisted Landscape — Land (Modern Horizons 3, #5948):
//
//	"{T}: Add {C}.
//	 {T}, Sacrifice this land: Search your library for a basic Swamp,
//	 Mountain, or Forest card, put it onto the battlefield tapped, then
//	 shuffle.
//	 Cycling {B}{R}{G}"
//
// A row of the Landscape cycle (s82GapB58Landscape). No simplification.
func init() {
	Register(s82GapB58Landscape("db659cae-2078-423e-a6ed-63898dbab87f", "Twisted Landscape", "{B}{R}{G}",
		"Swamp", "Mountain", "Forest"))
}
