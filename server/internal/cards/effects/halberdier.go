package effects

// Halberdier — Creature — Human Barbarian {3}{R}, 3/1 (Odyssey, #5613):
//
//	"First strike"
//
// First strike is an engine keyword: it deals its combat damage in the
// first-strike damage step (CR 702.7b).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "1c3c23b5-b771-4117-83b5-febf7e96281a",
		Name:            "Halberdier",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"first strike"},
	})
}
