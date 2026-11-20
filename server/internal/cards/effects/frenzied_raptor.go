package effects

// Frenzied Raptor — Creature — Dinosaur {2}{R}, 4/2 (Ixalan, #5532):
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
		OracleID:     "655c59d5-8666-476c-b494-411fa2e243a7",
		Name:         "Frenzied Raptor",
		Completeness: CompletenessFull,
	})
}
