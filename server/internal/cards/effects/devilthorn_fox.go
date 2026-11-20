package effects

// Devilthorn Fox — Creature — Fox {1}{W}, 3/1
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
		OracleID:     "8f80a7be-3da3-4e84-82a8-dd5d4b2c3f32",
		Name:         "Devilthorn Fox",
		Completeness: CompletenessFull,
	})
}
