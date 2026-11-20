package effects

// Tree Monkey — Creature — Monkey {G}, 1/1 (Portal Second Age,
// #5625):
//
//	"Reach (This creature can block creatures with flying.)"
//
// Reach is an engine keyword carried as printed data: it can block a
// creature with flying (CR 702.17b).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "d7318464-e400-424f-ae6d-a9070c64df04",
		Name:            "Tree Monkey",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"reach"},
	})
}
