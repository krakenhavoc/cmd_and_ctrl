package effects

// Wall of Opposition — Creature — Wall {3}{R}{R}, 0/6 (Legends, #5639):
//
//	"Defender (This creature can't attack.)
//	 {1}: This creature gets +1/+0 until end of turn."
//
// The shared legSelfPump row: no tap in the cost, so it can be activated
// any number of times, summoning sick or not, and the boost ends in the
// cleanup step (CR 514.2). Defender keeps it from attacking (CR 702.3b).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "30254aa2-2307-4f42-926a-54a416f55f87",
		Name:            "Wall of Opposition",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"defender"},
		Activated:       []ActivatedAbility{legSelfPump("{1}", 0, "This creature", "Wall of Opposition")},
	})
}
