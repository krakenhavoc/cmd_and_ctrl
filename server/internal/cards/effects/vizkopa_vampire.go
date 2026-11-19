package effects

// Vizkopa Vampire — Creature — Vampire {2}{W/B}, 3/1 (Ravnica: Clue Edition):
//
//	"Lifelink"
//
// The keyword rides PrintedKeywords; there is no other text.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "fcedc288-5e27-45a2-a07b-06c7aa7a9230",
		Name:            "Vizkopa Vampire",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"lifelink"},
	})
}
