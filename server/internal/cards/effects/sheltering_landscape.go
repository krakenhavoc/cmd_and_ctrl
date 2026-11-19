package effects

// Sheltering Landscape — Land (Modern Horizons 3, #5948):
//
//	"{T}: Add {C}.
//	 {T}, Sacrifice this land: Search your library for a basic Mountain,
//	 Forest, or Plains card, put it onto the battlefield tapped, then
//	 shuffle.
//	 Cycling {R}{G}{W}"
//
// A row of the Landscape cycle (s82GapB58Landscape). No simplification.
func init() {
	Register(s82GapB58Landscape("5b932be0-4dac-41b9-9c59-f79e4cecc31a", "Sheltering Landscape", "{R}{G}{W}",
		"Mountain", "Forest", "Plains"))
}
