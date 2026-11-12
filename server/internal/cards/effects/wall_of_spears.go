package effects

// Wall of Spears — Artifact Creature — Wall {3}, 2/3 (Antiquities, #5640):
//
//	"Defender
//	 First strike"
//
// Two printed keywords and nothing else: they ride PrintedKeywords, and
// the combat engine already refuses a defender as an attacker and deals
// first-strike damage in its own step.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "ed836d84-ff1e-4af8-b4b8-314569b3faec",
		Name:            "Wall of Spears",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"defender", "first strike"},
	})
}
