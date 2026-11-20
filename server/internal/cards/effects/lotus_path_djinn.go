package effects

// Lotus Path Djinn — Creature — Djinn Monk {3}{U}, 2/3 (Fate Reforged,
// #5552):
//
//	"Flying
//	 Prowess (Whenever you cast a noncreature spell, this creature gets
//	 +1/+1 until end of turn.)"
//
// Wing Commando's two keywords: flying, and prowess, which the engine builds
// as the CR 702.108a trigger.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "84dc0aae-029b-48dc-b8d0-e93c227e32e5",
		Name:            "Lotus Path Djinn",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying", "prowess"},
	})
}
