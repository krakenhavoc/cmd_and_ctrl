package effects

// Plated Spider — Creature — Spider {4}{G}, 4/4 (Urza's Destiny, #5621):
//
//	"Reach"
//
// One printed keyword, which the block check reads: it can block
// creatures with flying (CR 702.17b).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "15163c99-9388-4225-b707-4e057b2adfbc",
		Name:            "Plated Spider",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"reach"},
	})
}
