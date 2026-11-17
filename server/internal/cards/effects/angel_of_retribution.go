package effects

// Angel of Retribution — Creature — Angel {6}{W}, 5/5 (Torment, #5612):
//
//	"Flying, first strike"
//
// Two printed keywords the engine enforces: flying in block legality
// (CR 702.9b) and first strike in the combat damage steps (CR 702.7b).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "248dcf69-b99e-4a71-b9b3-7fd865f873ab",
		Name:            "Angel of Retribution",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying", "first strike"},
	})
}
