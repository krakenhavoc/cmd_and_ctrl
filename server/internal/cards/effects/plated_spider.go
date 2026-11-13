package effects

// Plated Spider — Creature — Spider {4}{G}, 4/4 (Urza's Destiny, #5621):
//
//	"Reach (This creature can block creatures with flying.)"
//
// Giant Spider's reach on a 4/4 (CR 702.17a): the block check lets it
// block a creature with flying.
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
