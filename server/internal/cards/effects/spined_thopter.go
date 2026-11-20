package effects

// Spined Thopter — Artifact Creature — Phyrexian Thopter {2}{U/P}, 2/1 (New Phyrexia, #5575):
//
//	"({U/P} can be paid with either {U} or 2 life.)
//	 Flying"
//
// The Phyrexian symbol is paid by the engine at cast (CR 107.4f); flying
// is the engine's (CR 702.9).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "45f2af87-0235-49f8-a63d-370faff3a097",
		Name:            "Spined Thopter",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying"},
	})
}
