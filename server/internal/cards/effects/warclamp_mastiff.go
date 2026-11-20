package effects

// Warclamp Mastiff — Creature — Dog {W}, 1/1 (Magic 2013, #5568):
//
//	"First strike (This creature deals combat damage before creatures
//	 without first strike.)"
//
// A keyword body: first strike (CR 702.7).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "8bd301ea-32f0-4c3a-a215-c04b31fbdc7c",
		Name:            "Warclamp Mastiff",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"first strike"},
	})
}
