package effects

// Ridgetop Raptor — Creature — Dinosaur Beast {3}{R}, 2/1 (Dominaria
// Remastered, #5948):
//
//	"Double strike"
//
// The keyword rides PrintedKeywords (CR 702.4).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "600a8c9e-e158-4fff-8c4d-c3a1bfd006e8",
		Name:            "Ridgetop Raptor",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"double strike"},
	})
}
