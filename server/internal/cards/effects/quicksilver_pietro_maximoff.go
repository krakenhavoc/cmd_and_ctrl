package effects

// Quicksilver, Pietro Maximoff — Legendary Creature — Mutant Hero {2}{R},
// 3/2 (Marvel Super Heroes Commander, #4455): "Haste". A keyword-only
// body, the same shape as Glamorous Grapplers.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "c5c1fb0f-b9cf-48a0-a4ea-e995e222e4b4",
		Name:            "Quicksilver, Pietro Maximoff",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"haste"},
	})
}
