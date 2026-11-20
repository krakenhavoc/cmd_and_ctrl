package effects

// Minotaur Aggressor — Creature — Minotaur Berserker {6}{R}, 6/2 (Return to Ravnica, #5566):
//
//	"First strike, haste"
//
// The printed keywords ride PrintedKeywords; the engine enforces them.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "7576bfba-eb4f-4d3e-aec5-52daee84d1d6",
		Name:            "Minotaur Aggressor",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"first strike", "haste"},
	})
}
