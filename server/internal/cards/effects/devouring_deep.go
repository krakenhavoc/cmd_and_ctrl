package effects

// Devouring Deep — Creature — Fish {2}{U}, 1/2 (Legends, #5639):
//
//	"Islandwalk (This creature can't be blocked as long as defending player controls an Island.)"
//
// Landwalk is enforced at the block declaration: it can't be blocked
// while the defending player controls an Island (CR 702.14c).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "e0f18302-5ccf-47d2-a8a1-bb7aeaca9a53",
		Name:            "Devouring Deep",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"islandwalk"},
	})
}
