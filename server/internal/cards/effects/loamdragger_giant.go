package effects

// Loamdragger Giant — Creature — Giant Warrior {4}{R/G}{R/G}{R/G}, 7/6
// (Shadowmoor, #5591):
//
//	(no rules text)
//
// A vanilla body, registered so the catalog knows it (see
// alaborn_trooper.go). The cost, types and body ride game.Card from
// Scryfall, so the Spec is the key and the name.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "b7523824-8c3e-4735-8fbc-6a5f99d1e1ac",
		Name:         "Loamdragger Giant",
		Completeness: CompletenessFull,
	})
}
