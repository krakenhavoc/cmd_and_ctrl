package effects

// Raging Redcap — Creature — Goblin Knight {2}{R}, 1/2 (Foundations, #4465):
//
//	"Double strike"
//
// Double strike is the canonical keyword token.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "679841ad-f3ac-48e2-ba11-2cf6d78642b4",
		Name:            "Raging Redcap",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"double strike"},
	})
}
