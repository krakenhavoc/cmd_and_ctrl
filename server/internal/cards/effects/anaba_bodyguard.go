package effects

// Anaba Bodyguard — Creature — Minotaur {3}{R}, 2/3 (Homelands, #5635):
//
//	"First strike"
//
// First strike is an engine keyword carried as printed data: the
// creature deals its combat damage in the first damage step (CR 702.7b).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "8c64db66-879e-46ad-9d11-8a95b6cf5d19",
		Name:            "Anaba Bodyguard",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"first strike"},
	})
}
