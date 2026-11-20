package effects

// Vigilant Baloth — Creature — Beast {3}{G}{G}, 5/5 (Core Set 2019, #5527):
//
//	"Vigilance (Attacking doesn't cause this creature to tap.)"
//
// The whole text is printed keywords; the engine reads them off
// PrintedKeywords.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "caf1bc69-a8d7-454b-9cc6-2e4d003776a8",
		Name:            "Vigilant Baloth",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"vigilance"},
	})
}
