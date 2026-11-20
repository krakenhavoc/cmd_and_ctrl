package effects

// Hovermyr — Artifact Creature — Myr {2}, 1/2 (New Phyrexia, #5575):
//
//	"Flying, vigilance"
//
// Both are engine keywords (CR 702.9b, CR 702.20b), as on Zephyr Falcon.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "b58722c5-84f2-43c2-be7e-3577c2f2a60a",
		Name:            "Hovermyr",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying", "vigilance"},
	})
}
