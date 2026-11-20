package effects

// Skyblade of the Legion — Creature — Vampire Soldier {1}{W}, 1/3
// (Ixalan, #5532):
//
//	"Flying"
//
// Flying is an engine keyword carried as printed data: it can be
// blocked only by creatures with flying or reach (CR 702.9b).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "678c76bf-fc31-47e4-8c19-712e15bba9e3",
		Name:            "Skyblade of the Legion",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying"},
	})
}
