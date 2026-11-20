package effects

// Kindercatch — Creature — Spirit {3}{G}{G}{G}, 6/6 (Innistrad, #5572):
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
		OracleID:     "c690f92f-f34f-4df0-9e01-d2ecf3b8f857",
		Name:         "Kindercatch",
		Completeness: CompletenessFull,
	})
}
