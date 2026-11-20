package effects

// Porcelain Legionnaire — Artifact Creature — Phyrexian Soldier {2}{W/P}, 3/1 (New Phyrexia, #5575):
//
//	"({W/P} can be paid with either {W} or 2 life.)
//	 First strike"
//
// The Phyrexian symbol is paid by the engine at cast (CR 107.4f); first
// strike is the engine's (CR 702.7).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "66442396-f804-47f3-8daa-0d43d8cfb482",
		Name:            "Porcelain Legionnaire",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"first strike"},
	})
}
