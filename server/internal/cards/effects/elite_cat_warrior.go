package effects

// Elite Cat Warrior — Creature — Cat Warrior {2}{G}, 2/3 (Portal, #5630):
//
//	"Forestwalk"
//
// Forestwalk is an engine keyword carried as printed data: it can't be
// blocked while the defending player controls a Forest (CR 702.14c).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "9297e50f-f8a5-4793-b64d-a119404f94ce",
		Name:            "Elite Cat Warrior",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"forestwalk"},
	})
}
