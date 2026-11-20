package effects

// Nimble Birdsticker — Creature — Goblin {2}{R}, 2/3 (Core Set 2020, #5518):
//
//	"Reach (This creature can block creatures with flying.)"
//
// A printed keyword.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "2f047fcb-c198-44ac-aa1e-31447f6c6aba",
		Name:            "Nimble Birdsticker",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"reach"},
	})
}
