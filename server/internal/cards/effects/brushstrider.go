package effects

// Brushstrider — Creature — Beast {1}{G}, 3/1 (Foundations Jumpstart):
//
//	"Vigilance"
//
// Vigilance is the engine keyword.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "6dfcc4bb-7cd5-4b3e-ad79-3c246a3af2a8",
		Name:            "Brushstrider",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"vigilance"},
	})
}
