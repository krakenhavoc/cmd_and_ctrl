package effects

// Goblin Striker — Creature — Goblin Berserker {1}{R}, 1/1 (Mirrodin,
// #5606):
//
//	"First strike, haste"
//
// Two printed keywords the engine already honours: first strike
// (CR 702.7b) and haste (CR 702.10b).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "c991d1b8-3adb-4854-9fd3-83f06aeb3941",
		Name:            "Goblin Striker",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"first strike", "haste"},
	})
}
