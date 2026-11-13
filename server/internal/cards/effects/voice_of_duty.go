package effects

// Voice of Duty — Creature — Angel {3}{W}, 2/2 (Urza's Destiny, #5621):
//
//	"Flying, protection from green"
//
// Treetop Sentinel's keywords on a white 2/2: green spells and
// abilities from green sources can't target it, and the rest of
// protection is the engine's (CR 702.16b).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "b5d5b9c9-accc-4dbe-aaf4-c806f3e1c61c",
		Name:            "Voice of Duty",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying", "protection from green"},
	})
}
