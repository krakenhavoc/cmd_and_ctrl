package effects

// Aven Envoy — Creature — Bird Soldier {U}, 0/2 (Legions, #5609):
//
//	"Flying"
//
// One printed keyword (CR 702.9b).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "73cd8461-7930-4e2d-9a14-4d10db91e93e",
		Name:            "Aven Envoy",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying"},
	})
}
