package effects

// Starlit Angel — Creature — Angel {3}{W}{W}, 3/4 (Portal, #5630):
//
//	"Flying"
//
// Flying is an engine keyword carried as printed data: it can be blocked
// only by creatures with flying or reach (CR 702.9b).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "e66e3450-bddb-46ac-bc7e-c7732e258374",
		Name:            "Starlit Angel",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying"},
	})
}
