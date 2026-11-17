package effects

// Tempest Drake — Creature — Drake {1}{W}{U}, 2/2 (Visions, #5631):
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
		OracleID:        "ff0086e0-706f-4474-9b5e-a1591235bf9b",
		Name:            "Tempest Drake",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying", "vigilance"},
	})
}
