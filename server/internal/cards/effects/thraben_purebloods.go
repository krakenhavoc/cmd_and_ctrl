package effects

// Thraben Purebloods — Creature — Dog {4}{W}, 3/5 (Innistrad, #5572):
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
		OracleID:     "cf8aef54-3f43-413c-9f3f-d4a1b2090904",
		Name:         "Thraben Purebloods",
		Completeness: CompletenessFull,
	})
}
