package effects

// Righteous Avengers — Creature — Human Soldier {4}{W}, 3/1 (Legends, #5639):
//
//	"Plainswalk (This creature can't be blocked as long as defending player controls a Plains.)"
//
// Landwalk is enforced at the block declaration: it can't be blocked
// while the defending player controls a Plains (CR 702.14c).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "e7c07e41-3654-47da-afe3-88aae4c034b6",
		Name:            "Righteous Avengers",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"plainswalk"},
	})
}
