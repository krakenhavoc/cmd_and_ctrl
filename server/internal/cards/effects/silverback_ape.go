package effects

// Silverback Ape — Creature — Ape {3}{G}{G}, 5/5
// (Starter 1999, #5620):
//
//	(no rules text)
//
// A vanilla body, registered so the catalog knows it (Grizzly Bears'
// shape). The cost, types and body ride game.Card from Scryfall, so the
// Spec is the key and the name.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "a72f9437-d652-4a64-99df-a307a6dd9a0d",
		Name:         "Silverback Ape",
		Completeness: CompletenessFull,
	})
}
