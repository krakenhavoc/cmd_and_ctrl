package effects

// Ramirez DePietro — Legendary Creature — Human Pirate {3}{U}{B}{B}, 4/3 (Legends, #5639):
//
//	"First strike"
//
// First strike is an engine keyword: it deals its combat damage in the
// first-strike damage step (CR 702.7b).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "48b9c510-03ed-41cb-8c17-16110d447490",
		Name:            "Ramirez DePietro",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"first strike"},
	})
}
