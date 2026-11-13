package effects

// Willow Dryad — Creature — Dryad {G}, 1/1 (Portal, #5630):
//
//	"Forestwalk"
//
// Forestwalk is an engine keyword carried as printed data: it can't be
// blocked while the defending player controls a Forest (CR 702.14c).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "19d5eb29-2d35-44e7-afa1-3ba8d812ed41",
		Name:            "Willow Dryad",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"forestwalk"},
	})
}
