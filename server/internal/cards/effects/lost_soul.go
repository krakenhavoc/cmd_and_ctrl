package effects

// Lost Soul — Creature — Spirit Minion {1}{B}{B}, 2/1 (Legends, #5639):
//
//	"Swampwalk (This creature can't be blocked as long as defending player controls a Swamp.)"
//
// Landwalk is enforced at the block declaration: it can't be blocked
// while the defending player controls a Swamp (CR 702.14c).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "1c68e877-ee47-42bf-a867-3f7f479eae1c",
		Name:            "Lost Soul",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"swampwalk"},
	})
}
