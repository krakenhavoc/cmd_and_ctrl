package effects

// Harrier Naga — Creature — Snake Warrior {2}{G}, 3/3 (Hour of Devastation, #5534):
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
		OracleID:     "4a1cbac1-ee22-44cf-a37f-a58055f5c9bc",
		Name:         "Harrier Naga",
		Completeness: CompletenessFull,
	})
}
