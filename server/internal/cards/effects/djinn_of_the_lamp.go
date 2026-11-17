package effects

// Djinn of the Lamp — Creature — Djinn {5}{U}{U}, 5/6 (Portal, #5630):
//
//	"Flying"
//
// Flying is an engine keyword carried as printed data: it can be blocked
// only by creatures with flying or reach (CR 702.9b).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "72b42c63-fe4d-4823-9692-30fb5bab384a",
		Name:            "Djinn of the Lamp",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying"},
	})
}
