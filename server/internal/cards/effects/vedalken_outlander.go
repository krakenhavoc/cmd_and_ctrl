package effects

// Vedalken Outlander — Artifact Creature — Vedalken Scout {W}{U}, 2/2 (Conflux,
// #5587):
//
//	"Protection from red"
//
// Protection from red is a printed keyword, read by targeting, blocking,
// damage and attaching (CR 702.16a).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "814a30dc-14e4-45cb-9320-b043e686b8ba",
		Name:            "Vedalken Outlander",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"protection from red"},
	})
}
