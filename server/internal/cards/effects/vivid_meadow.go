package effects

// Vivid Meadow — Land (Vivid cycle, S82 reprint pass):
//
//	"This land enters tapped with two charge counters on it.
//	 {T}: Add {W}.
//	 {T}, Remove a charge counter from this land: Add one mana of any color."
//
// The cycle's shared body (vivid_lands.go).
//
// No simplification.
func init() {
	Register(vividLand("dee99df5-628f-4a4e-a203-4dfddc927373", "Vivid Meadow", "W"))
}
