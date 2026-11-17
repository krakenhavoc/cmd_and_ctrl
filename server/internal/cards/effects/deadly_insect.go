package effects

// Deadly Insect — Creature — Insect {4}{G}, 6/1 (Alliances, #5634):
//
//	"Shroud"
//
// Shroud is an engine keyword carried as printed data: no spell or
// ability can target it, its controller's included (CR 702.18a).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "98273b0d-2b41-486b-9498-30a692c03982",
		Name:            "Deadly Insect",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"shroud"},
	})
}
