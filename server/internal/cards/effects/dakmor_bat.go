package effects

// Dakmor Bat — Creature — Bat {1}{B}, 1/1 (Portal Second Age,
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
		OracleID:        "724d8247-33b1-42bf-a827-7b66e45504c1",
		Name:            "Dakmor Bat",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying"},
	})
}
