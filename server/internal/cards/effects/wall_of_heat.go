package effects

// Wall of Heat — Creature — Wall {2}{R}, 2/6 (Legends, #5639):
//
//	"Defender (This creature can't attack.)"
//
// Defender is an engine keyword: the creature can't be declared as an
// attacker (CR 702.3b).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "c7bbfd32-2a6f-44e9-b51e-9b272f48d458",
		Name:            "Wall of Heat",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"defender"},
	})
}
