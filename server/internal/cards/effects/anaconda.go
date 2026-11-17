package effects

// Anaconda — Creature — Snake {3}{G}, 3/3 (Portal, #5630):
//
//	"Swampwalk"
//
// Swampwalk is an engine keyword carried as printed data: it can't be
// blocked while the defending player controls a Swamp (CR 702.14c).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "3eff03f1-2c5f-4c59-b465-a8c4cd05e1ba",
		Name:            "Anaconda",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"swampwalk"},
	})
}
