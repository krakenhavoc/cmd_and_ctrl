package effects

// Hired Blade — Creature — Human Assassin {2}{B}, 3/2 (Assassin's Creed,
// #5948):
//
//	"Flash"
//
// A vanilla 3/2 with flash: the keyword lets it be cast any time its
// controller could cast an instant (CR 702.8a).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "fed32c9c-7458-4143-b0b2-45ca23ba25ee",
		Name:            "Hired Blade",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flash"},
	})
}
