package effects

// Fire Elemental — Creature — Elemental {3}{R}{R}, 5/4 (Foundations,
// #4465):
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
		OracleID:     "3912d21e-1ebc-4a81-9dc9-f404248d564a",
		Name:         "Fire Elemental",
		Completeness: CompletenessFull,
	})
}
