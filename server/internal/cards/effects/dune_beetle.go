package effects

// Dune Beetle — Creature — Insect {1}{B}, 1/4
// (Amonkhet, #5535):
//
//	(no rules text)
//
// A vanilla body, registered so the catalog knows it (see
// grizzly_bears.go). The cost, types and body ride game.Card from
// Scryfall, so the Spec is the key and the name.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "3f79603a-3173-404e-8c69-b2254c1270dc",
		Name:         "Dune Beetle",
		Completeness: CompletenessFull,
	})
}
