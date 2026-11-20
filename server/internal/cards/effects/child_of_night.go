package effects

// Child of Night — Creature — Vampire {1}{B}, 2/1 (Magic 2010, #5585):
//
//	"Lifelink"
//
// Lifelink is an engine keyword carried as printed data (CR 702.15b).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "c650a7bc-e350-44a0-a698-d4a233d66156",
		Name:            "Child of Night",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"lifelink"},
	})
}
