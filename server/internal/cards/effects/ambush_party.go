package effects

// Ambush Party — Creature — Human Rogue {4}{R}, 3/1 (Homelands, #5635):
//
//	"First strike, haste"
//
// Both are engine keywords carried as printed data: first strike puts
// its combat damage in the first damage step (CR 702.7b), and haste
// lets it attack the turn it arrives (CR 702.10b).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "454ed336-362d-46d9-96d5-8fcaa8a1c543",
		Name:            "Ambush Party",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"first strike", "haste"},
	})
}
