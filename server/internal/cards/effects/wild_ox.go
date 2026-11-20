package effects

// Wild Ox — Creature — Ox {3}{G}, 3/3 (Portal Second Age,
// #5625):
//
//	"Swampwalk (This creature can't be blocked as long as defending player controls a Swamp.)"
//
// Swampwalk is an engine keyword carried as printed data: it can't be
// blocked while the defending player controls a Swamp (CR 702.14c).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "a6aa245c-b620-486c-864b-0ea4e9bc4048",
		Name:            "Wild Ox",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"swampwalk"},
	})
}
