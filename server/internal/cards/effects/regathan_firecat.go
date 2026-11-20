package effects

// Regathan Firecat — Creature — Elemental Cat {2}{R}, 4/1 (Magic 2014, #5563):
//
//	(no rules text)
//
// A vanilla body, registered so the catalog knows it (see ordinary_bear.go).
// The cost, types and body ride game.Card from Scryfall, so the Spec is the key
// and the name.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "b3f609b5-fb4e-4e66-a70a-53bbe0fa2b9a",
		Name:         "Regathan Firecat",
		Completeness: CompletenessFull,
	})
}
