package effects

// Moaning Spirit — Creature — Spirit {2}{B}, 2/1 (Portal Second Age,
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
		OracleID:        "78962897-e2e3-4cb2-ae97-eca5dbe74e49",
		Name:            "Moaning Spirit",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying"},
	})
}
