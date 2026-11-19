package effects

// Seething Landscape — Land (Modern Horizons 3, #5948):
//
//	"{T}: Add {C}.
//	 {T}, Sacrifice this land: Search your library for a basic Island,
//	 Swamp, or Mountain card, put it onto the battlefield tapped, then
//	 shuffle.
//	 Cycling {U}{B}{R}"
//
// A row of the Landscape cycle (s82GapB58Landscape). No simplification.
func init() {
	Register(s82GapB58Landscape("2d8635bd-ed96-4bb1-8718-6962a0eee3d5", "Seething Landscape", "{U}{B}{R}",
		"Island", "Swamp", "Mountain"))
}
