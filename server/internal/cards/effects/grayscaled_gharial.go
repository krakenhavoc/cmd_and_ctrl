package effects

// Grayscaled Gharial — Creature — Crocodile {U}, 1/1 (Ravnica: City of
// Guilds, #5600):
//
//	"Islandwalk (This creature can't be blocked as long as defending
//	 player controls an Island.)"
//
// Islandwalk is an engine landwalk keyword carried as printed data (CR
// 702.14).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "59483597-4d8c-4cdc-8bb5-226b1eef10f5",
		Name:            "Grayscaled Gharial",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"islandwalk"},
	})
}
