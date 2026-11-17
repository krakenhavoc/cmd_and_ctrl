package effects

// Midnight Assassin (Streets of New Capenna, #5491):
//
//	Flying, deathtouch
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "f394d2bf-89aa-44c1-b473-8c7a8874a9a3",
		Name:            "Midnight Assassin",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying", "deathtouch"},
	})
}
