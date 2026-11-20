package effects

// Vampire Noble — Creature — Vampire Noble {2}{B}, 3/2
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
		OracleID:     "e85d957b-35a8-48a9-8d63-90e3c680e0dd",
		Name:         "Vampire Noble",
		Completeness: CompletenessFull,
	})
}
