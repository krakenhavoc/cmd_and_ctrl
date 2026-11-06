package effects

// Worf, Chief Tactical Officer — Legendary Creature — Klingon Warrior
// Officer {2}{W}, 2/3 (Star Trek Commander, #4974):
//
//	"Vigilance"
//
// The printed keyword, which the engine enforces (CR 702.20b).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "26e86be4-52e1-47cb-b548-c434482bd3ad",
		Name:            "Worf, Chief Tactical Officer",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"vigilance"},
	})
}
