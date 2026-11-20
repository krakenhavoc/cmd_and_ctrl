package effects

// Zodiac Ox — Creature — Ox {3}{G}, 3/3 (Portal Three Kingdoms, #5622):
//
//	"Swampwalk (This creature can't be blocked as long as defending player controls a Swamp.)"
//
// Swampwalk is an engine keyword carried as printed data: it can't be
// blocked while the defending player controls a Swamp (CR 702.14c).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "9c9e71d9-a636-4574-8b7e-e6dd0f71c7c8",
		Name:            "Zodiac Ox",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"swampwalk"},
	})
}
