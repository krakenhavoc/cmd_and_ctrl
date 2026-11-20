package effects

// Murmuring Phantasm — Creature — Spirit {1}{U}, 0/5:
//
//	"Defender"
//
// Defender, read at the attack declaration (CR 702.3b).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "35e684fd-69e3-4137-bf15-94f705191f5e",
		Name:            "Murmuring Phantasm",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"defender"},
	})
}
