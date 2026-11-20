package effects

// Kelinore Bat — Creature — Bat {2}{B}, 2/1 (Magic 2010, #5585):
//
//	"Flying"
//
// Flying is an engine keyword carried as printed data (CR 702.9b).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "552dcf9f-ff4a-4057-a93e-62f5527050c9",
		Name:            "Kelinore Bat",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying"},
	})
}
