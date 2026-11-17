package effects

// Archangel — Creature — Angel {5}{W}{W}, 5/5 (Visions, #5631):
//
//	"Flying, vigilance"
//
// Both are engine keywords carried as printed data: flying limits what
// can block it (CR 702.9b), and attacking does not tap a creature with
// vigilance (CR 702.20b).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "9971697b-2acc-4bc2-a44e-074d03a51df7",
		Name:            "Archangel",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying", "vigilance"},
	})
}
