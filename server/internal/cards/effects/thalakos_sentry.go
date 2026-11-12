package effects

// Thalakos Sentry — Creature — Thalakos Soldier {1}{U}, 1/2
// (Tempest, #5628):
//
//	"Shadow (This creature can block or be blocked by only creatures
//	 with shadow.)"
//
// Soltari Foot Soldier's shape: shadow is a printed keyword, and the
// block rules enforce it both ways, so the Sentry can block only an
// attacker with shadow (CR 702.28b).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "ca65b908-e6fe-4ca3-8ecd-247df957aa0f",
		Name:            "Thalakos Sentry",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"shadow"},
	})
}
