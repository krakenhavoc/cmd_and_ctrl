package effects

// Coral Merfolk — Creature — Merfolk {1}{U}, 2/1 (Urza's Saga, #5624):
//
//	(no rules text)
//
// A vanilla creature, registered for the reason Caelorna, Coral Tyrant
// gives: a card the catalog does not know falls back to the manual
// sandbox path, so even a body with no text is registered to resolve and
// enter on its own. The cost, the type and the 2/1 ride the card from
// Scryfall.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "4ed27607-21a8-4bc3-997e-6d2242313f6d",
		Name:         "Coral Merfolk",
		Completeness: CompletenessFull,
	})
}
