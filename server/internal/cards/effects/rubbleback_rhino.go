package effects

// Rubbleback Rhino — Creature — Rhino {4}{G}, 3/4 (Return to Ravnica, #5566):
//
//	"Hexproof (This creature can't be the target of spells or
//	 abilities your opponents control.)"
//
// The printed keywords ride PrintedKeywords; the engine enforces them.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "33e87ab4-f680-4463-a48d-289fed7e196e",
		Name:            "Rubbleback Rhino",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"hexproof"},
	})
}
