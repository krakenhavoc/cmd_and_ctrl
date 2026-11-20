package effects

// Vivid Crag — Land (Vivid cycle, Lorwyn, #5593):
//
//	"This land enters tapped with two charge counters on it.
//	 {T}: Add {R}.
//	 {T}, Remove a charge counter from this land: Add one mana of any
//	 color."
//
// The cycle's shared vividLand (vivid_lands.go).
//
// No simplification.
func init() {
	Register(vividLand("d8e2efe0-33a4-4303-9e83-ac42ea5df8cb", "Vivid Crag", "R"))
}
