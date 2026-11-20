package effects

// Chapel Geist — Creature — Spirit {1}{W}{W}, 2/3 (Innistrad, #5572):
//
//	"Flying"
//
// Flying is an engine keyword carried as printed data: only a creature
// with flying or reach can block it (CR 702.9b).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "f8219a99-1620-40aa-a257-fc3dfb239d51",
		Name:            "Chapel Geist",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying"},
	})
}
