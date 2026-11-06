package effects

// Caelorna, Coral Tyrant — Legendary Creature — Octopus {1}{U}, 0/8
// (Aetherdrift, #4464):
//
//	(no rules text)
//
// Another of the set's legendary vanillas, registered for the reason
// Sundial, Dawn Tyrant gives: a card the catalog does not know falls back
// to the manual sandbox path, so even a body with no text is registered to
// resolve and enter on its own. The cost, the types and the 0/8 ride the
// card from Scryfall; the legend rule is the engine's (CR 704.5j).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "4bbc5df9-60d7-46c1-8856-91c9c74abf18",
		Name:         "Caelorna, Coral Tyrant",
		Completeness: CompletenessFull,
	})
}
