package effects

// Zephyr Sprite — Creature — Faerie {U}, 1/1 (Magic 2010, #5585):
//
//	"Flying"
//
// Flying is an engine keyword carried as printed data (CR 702.9b).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "11f77158-0240-4882-94a4-6bd6b3d94bcc",
		Name:            "Zephyr Sprite",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying"},
	})
}
