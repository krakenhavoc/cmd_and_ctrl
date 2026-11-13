package effects

// Narwhal — Creature — Whale {2}{U}{U}, 2/2 (Homelands, #5635):
//
//	"First strike, protection from red"
//
// Both are engine keywords carried as printed data: first strike puts
// its combat damage in the first damage step (CR 702.7b), and
// protection from red stops red spells and abilities from red sources
// targeting it and prevents damage from red sources (CR 702.16b,
// CR 702.16e).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "3552eea9-aee9-4555-b3cb-8e4fc5b728cc",
		Name:            "Narwhal",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"first strike", "protection from red"},
	})
}
