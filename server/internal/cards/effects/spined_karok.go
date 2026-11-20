package effects

// Spined Karok — Creature — Crocodile {2}{G}, 2/4 (Strixhaven: School of
// Mages, #5502):
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
		OracleID:     "7341a04a-b6d6-4106-b1c3-3483f9d37663",
		Name:         "Spined Karok",
		Completeness: CompletenessFull,
	})
}
