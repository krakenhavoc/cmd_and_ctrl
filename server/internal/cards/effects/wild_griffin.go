package effects

// Wild Griffin — Creature — Griffin {2}{W}, 2/2 (Portal Second Age,
// #5625):
//
//	"Flying"
//
// Flying is an engine keyword carried as printed data: it can be blocked
// only by creatures with flying or reach (CR 702.9b).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "c643cfe1-5844-4eb1-b1f5-028382411773",
		Name:            "Wild Griffin",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying"},
	})
}
