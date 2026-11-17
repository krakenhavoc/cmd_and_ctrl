package effects

// Sabertooth Wyvern — Creature — Drake {4}{R}, 3/2 (Exodus, #5626):
//
//	"Flying, first strike"
//
// Two printed keywords, both enforced by the engine: flying in block
// legality (CR 702.9b) and first strike in the combat damage step
// (CR 702.7b).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "acaae632-3203-4867-ad03-32619c3fcfe6",
		Name:            "Sabertooth Wyvern",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying", "first strike"},
	})
}
