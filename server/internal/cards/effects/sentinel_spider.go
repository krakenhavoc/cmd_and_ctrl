package effects

// Sentinel Spider — Creature — Spider {3}{G}{G}, 4/4 (Magic 2013, #5568):
//
//	"Reach, vigilance"
//
// A keyword body: reach (CR 702.17) and vigilance (CR 702.20).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "a37c9b13-dc45-4b55-a63d-5314abde7aab",
		Name:            "Sentinel Spider",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"reach", "vigilance"},
	})
}
