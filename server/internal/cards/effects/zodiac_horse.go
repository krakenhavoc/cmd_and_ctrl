package effects

// Zodiac Horse — Creature — Horse {3}{G}, 3/3 (Portal Three Kingdoms, #5622):
//
//	"Islandwalk (This creature can't be blocked as long as defending player controls an Island.)"
//
// Islandwalk is an engine keyword carried as printed data: it can't be
// blocked while the defending player controls an Island (CR 702.14c).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "f778bf66-e823-4b4c-8586-97e1d519efd8",
		Name:            "Zodiac Horse",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"islandwalk"},
	})
}
