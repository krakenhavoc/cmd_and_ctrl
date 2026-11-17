package effects

// Storm Crow — Creature — Bird {1}{U}, 1/2 (Alliances, #5634):
//
//	"Flying"
//
// Flying is an engine keyword carried as printed data: it can be blocked
// only by creatures with flying or reach (CR 702.9b).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "000d5588-5a4c-434e-988d-396632ade42c",
		Name:            "Storm Crow",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying"},
	})
}
