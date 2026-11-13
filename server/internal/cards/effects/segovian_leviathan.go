package effects

// Segovian Leviathan — Creature — Leviathan {4}{U}, 3/3 (Legends, #5639):
//
//	"Islandwalk (This creature can't be blocked as long as defending player controls an Island.)"
//
// Landwalk is enforced at the block declaration: it can't be blocked
// while the defending player controls an Island (CR 702.14c).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "853b02e1-fc0a-413b-8381-88596624440f",
		Name:            "Segovian Leviathan",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"islandwalk"},
	})
}
