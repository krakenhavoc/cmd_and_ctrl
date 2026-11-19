package effects

// River Bear — Creature — Bear {3}{G}, 3/3 (Portal Second Age,
// #5625):
//
//	"Islandwalk (This creature can't be blocked as long as defending player controls an Island.)"
//
// Islandwalk is an engine keyword carried as printed data: it can't be
// blocked while the defending player controls an Island (CR 702.14c).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "568c4d02-501e-44ce-871c-adf9e4b9d8d6",
		Name:            "River Bear",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"islandwalk"},
	})
}
