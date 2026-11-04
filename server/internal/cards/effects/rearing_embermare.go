package effects

// Rearing Embermare — Creature — Horse Beast {4}{R}, 4/5 (Secrets of
// Strixhaven, #4456):
//
//	"Reach, haste"
//
// Both are printed keywords the engine reads off the card.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "95f20067-4471-4bfa-952d-dbbfbb2c2178",
		Name:            "Rearing Embermare",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"reach", "haste"},
	})
}
