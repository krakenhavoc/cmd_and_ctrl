package effects

// Azure Drake — Creature — Drake {3}{U}, 2/4 (Legends, #5639):
//
//	"Flying"
//
// Flying is an engine keyword: only a creature with flying or reach can
// block it (CR 702.9b).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "15cc068c-f424-433e-b165-8457c62a4b35",
		Name:            "Azure Drake",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying"},
	})
}
