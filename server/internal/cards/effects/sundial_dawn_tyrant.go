package effects

// Sundial, Dawn Tyrant — Legendary Artifact Creature — Construct {1}{W},
// 3/3 (Aetherdrift, #4464):
//
//	(no rules text)
//
// One of the set's legendary vanillas. Gigantosaurus's reason for an
// entry: a card the catalog does not know falls back to the manual
// sandbox path, so even a body with no text is registered to resolve and
// enter on its own. The cost, the types and the 3/3 ride the card from
// Scryfall; the legend rule is the engine's (CR 704.5j).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "9ca808c5-9338-4199-a45b-df1a1a20625e",
		Name:         "Sundial, Dawn Tyrant",
		Completeness: CompletenessFull,
	})
}
