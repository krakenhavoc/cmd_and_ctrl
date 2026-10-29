package effects

// Yargle, Glutton of Urborg — Legendary Creature — Frog Spirit {4}{B},
// 9/3:
//
//	(no rules text)
//
// Reprinted in Reality Fracture (FRA 241), the one card of that set the
// catalog was missing (#4453). A vanilla card still has to be
// REGISTERED to behave (see gigantosaurus.go): a card the catalog does
// not know falls back to the manual sandbox path. The cost, types and
// body ride game.Card from Scryfall, so the Spec is the key and the
// name.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "ed66cd31-958f-4b28-82a3-e04acc819afc",
		Name:         "Yargle, Glutton of Urborg",
		Completeness: CompletenessFull,
	})
}
