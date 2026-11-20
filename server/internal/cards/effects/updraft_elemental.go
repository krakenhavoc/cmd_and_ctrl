package effects

// Updraft Elemental — Creature — Elemental {2}{U}, 1/4 (Dragons of Tarkir,
// #5551):
//
//	"Flying"
//
// Flying is an engine keyword carried as printed data (CR 702.9).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "a33982e4-d7d2-4319-bf04-534b8f253dfb",
		Name:            "Updraft Elemental",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying"},
	})
}
