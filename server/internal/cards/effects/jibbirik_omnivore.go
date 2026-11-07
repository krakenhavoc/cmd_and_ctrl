package effects

// Jibbirik Omnivore — Creature — Beast {1}{G}, 3/2 (Aetherdrift, #4464):
//
//	(no rules text)
//
// A vanilla, registered for the reason Sundial, Dawn Tyrant gives: a card
// the catalog does not know falls back to the manual sandbox path, so even
// a body with no text is registered to resolve and enter on its own. The
// cost, the type and the 3/2 ride the card from Scryfall.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "9cda6cb8-c254-413a-babd-0e24e7b89cc4",
		Name:         "Jibbirik Omnivore",
		Completeness: CompletenessFull,
	})
}
