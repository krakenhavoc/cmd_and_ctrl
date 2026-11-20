package effects

// Haunted Guardian — {2} Artifact Creature — Construct 2/1 (Avacyn Restored, #5570):
//
//	"Defender, first strike"
//
// Two printed keywords (CR 702.3, CR 702.7).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "20b9bc53-7f6d-4616-a2df-5783ab577707",
		Name:            "Haunted Guardian",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"defender", "first strike"},
	})
}
