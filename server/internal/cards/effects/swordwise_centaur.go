package effects

// Swordwise Centaur — Creature — Centaur Warrior {G}{G}, 3/2 (Born of the Gods,
// #5559):
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
		OracleID:     "5c2e0762-4994-4c34-b33f-909c2b3f0f43",
		Name:         "Swordwise Centaur",
		Completeness: CompletenessFull,
	})
}
