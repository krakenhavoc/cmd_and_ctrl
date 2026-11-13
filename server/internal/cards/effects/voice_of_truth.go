package effects

// Voice of Truth — Creature — Angel {3}{W}, 2/2 (Nemesis, #5618):
//
//	"Flying, protection from white"
//
// Two printed keywords, both enforced by the engine. Only a creature
// with flying or reach can block it (CR 702.9b), and a white one can't
// (CR 702.16f); white spells and abilities can't target it (CR 702.16b),
// and damage from white sources is prevented (CR 702.16e).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "33435856-ea24-4d10-84d5-6872b35b63a3",
		Name:            "Voice of Truth",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying", "protection from white"},
	})
}
