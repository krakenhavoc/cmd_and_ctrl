package effects

// Wildwood Patrol — Creature — Centaur Scout {2}{G}, 4/2 (Ravnica: Clue Edition):
//
//	"Trample"
//
// The keyword rides PrintedKeywords; there is no other text.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "4fb9c040-c684-4ba8-970e-58313af5584a",
		Name:            "Wildwood Patrol",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"trample"},
	})
}
