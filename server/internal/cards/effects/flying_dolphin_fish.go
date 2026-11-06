package effects

// Flying Dolphin-Fish — Creature — Whale Fish {1}{U}, 1/3 (Avatar: The Last
// Airbender Eternal, #5075):
//
//	"Flying"
//
// Flying is an engine keyword, carried as printed data.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "59fa8323-bba9-4c51-833d-2ab966599b62",
		Name:            "Flying Dolphin-Fish",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying"},
	})
}
