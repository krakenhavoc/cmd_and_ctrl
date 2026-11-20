package effects

// Silent Observer — Creature — Spirit {3}{U}, 1/5
// (Shadows over Innistrad, #5543):
//
//	"Flying"
//
// Flying is an engine keyword carried as printed data (CR 702.9).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "c377c3e0-2b91-43c6-9f49-5dd6fbf94308",
		Name:            "Silent Observer",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying"},
	})
}
