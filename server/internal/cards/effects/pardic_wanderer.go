package effects

// Pardic Wanderer — Artifact Creature — Golem {6}, 5/5 (Dominaria, #5530):
//
//	"Trample"
//
// One printed keyword, read by combat damage assignment (CR 702.19b).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "82ff4500-b118-4ce7-9f7c-92661328b242",
		Name:            "Pardic Wanderer",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"trample"},
	})
}
