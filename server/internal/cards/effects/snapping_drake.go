package effects

// Snapping Drake — Creature — Drake {3}{U}, 3/2 (Portal, #5630):
//
//	"Flying"
//
// Flying is an engine keyword carried as printed data: it can be blocked
// only by creatures with flying or reach (CR 702.9b).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "e15060c3-3773-4548-8747-ff59dcf2b519",
		Name:            "Snapping Drake",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying"},
	})
}
