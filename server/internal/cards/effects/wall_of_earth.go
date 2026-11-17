package effects

// Wall of Earth — Creature — Wall {1}{R}, 0/6 (Legends, #5639):
//
//	"Defender (This creature can't attack.)"
//
// Defender is an engine keyword: the creature can't be declared as an
// attacker (CR 702.3b).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "1e3e1916-bb82-481e-9b10-225504ef06b3",
		Name:            "Wall of Earth",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"defender"},
	})
}
