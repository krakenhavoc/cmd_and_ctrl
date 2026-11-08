package effects

// Kalakscion, Hunger Tyrant — Legendary Creature — Crocodile {1}{B}{B},
// 7/2 (Aetherdrift, #4464):
//
//	(no rules text)
//
// Another of the set's legendary vanillas, registered for the reason
// Sundial, Dawn Tyrant gives: a card the catalog does not know falls back
// to the manual sandbox path, so even a body with no text is registered to
// resolve and enter on its own. The cost, the types and the 7/2 ride the
// card from Scryfall; the legend rule is the engine's (CR 704.5j).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "d85f5edb-60a8-4abd-82b0-f9a58f3d4185",
		Name:         "Kalakscion, Hunger Tyrant",
		Completeness: CompletenessFull,
	})
}
