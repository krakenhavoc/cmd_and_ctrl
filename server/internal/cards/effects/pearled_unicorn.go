package effects

// Pearled Unicorn — Creature — Unicorn {2}{W}, 2/2
// (Limited Edition Alpha, #5642):
//
//	(no rules text)
//
// A vanilla body, registered so the catalog knows it (see
// ordinary_bear.go). The cost, types and body ride game.Card from
// Scryfall, so the Spec is the key and the name.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "c071be90-0531-40cc-af46-0cbe80c4ddd4",
		Name:         "Pearled Unicorn",
		Completeness: CompletenessFull,
	})
}
