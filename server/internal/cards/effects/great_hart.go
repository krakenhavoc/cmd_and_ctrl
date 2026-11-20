package effects

// Great Hart — Creature — Elk {3}{W}, 2/4 (Born of the Gods,
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
		OracleID:     "694afc0a-bb32-4f24-9347-66665117ac9f",
		Name:         "Great Hart",
		Completeness: CompletenessFull,
	})
}
