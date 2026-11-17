package effects

// Thunder Spirit — Creature — Elemental Spirit {1}{W}{W}, 2/2 (Legends, #5639):
//
//	"Flying, first strike"
//
// Both are engine keywords: only a flyer or a creature with reach blocks
// it (CR 702.9b), and it deals its combat damage in the first-strike
// step (CR 702.7b).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "6efad3d6-4886-4b62-a4ff-bc0622f9672a",
		Name:            "Thunder Spirit",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying", "first strike"},
	})
}
