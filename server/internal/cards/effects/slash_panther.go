package effects

// Slash Panther — Artifact Creature — Phyrexian Cat {4}{R/P}, 4/2 (New Phyrexia, #5575):
//
//	"({R/P} can be paid with either {R} or 2 life.)
//	 Haste"
//
// The Phyrexian symbol is paid by the engine at cast (CR 107.4f); haste is
// the engine's (CR 702.10).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "76558381-c4ae-4bc3-8c7b-fa0b717f0282",
		Name:            "Slash Panther",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"haste"},
	})
}
