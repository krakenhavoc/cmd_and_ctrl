package effects

// Blanchwood Treefolk — Creature — Treefolk {4}{G}, 4/5 (Urza's Saga,
// #5624):
//
//	(no rules text)
//
// A vanilla creature, registered for the reason Caelorna, Coral Tyrant
// gives: a card the catalog does not know falls back to the manual
// sandbox path, so even a body with no text is registered to resolve and
// enter on its own. The cost, the type and the 4/5 ride the card from
// Scryfall.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "a3090daf-26d9-4ba5-9fa3-42bd128e229d",
		Name:         "Blanchwood Treefolk",
		Completeness: CompletenessFull,
	})
}
