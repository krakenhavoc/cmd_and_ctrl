package effects

// Goblin Trailblazer — Creature — Goblin Pirate {1}{R}, 2/1 (Foundations
// Jumpstart):
//
//	"Menace"
//
// A printed keyword. No simplification.
func init() {
	Register(Spec{
		OracleID:        "38f66ef9-6afc-478b-a1d4-a9d42c31dfc9",
		Name:            "Goblin Trailblazer",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"menace"},
	})
}
