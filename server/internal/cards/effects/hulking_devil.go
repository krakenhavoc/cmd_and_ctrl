package effects

// Hulking Devil — Creature — Devil {3}{R}, 5/2
// (Shadows over Innistrad, #5543):
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
		OracleID:     "b6578a36-c43c-4474-81bd-5711cc4f626e",
		Name:         "Hulking Devil",
		Completeness: CompletenessFull,
	})
}
