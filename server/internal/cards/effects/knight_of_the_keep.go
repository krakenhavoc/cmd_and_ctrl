package effects

// Knight of the Keep — Creature — Human Knight {2}{W}, 3/2 (Throne of Eldraine, #5516):
//
//	(no rules text)
//
// A vanilla body, registered so the catalog knows it (see grizzly_bears.go).
// The cost, types and body ride game.Card from Scryfall.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "ab1d04ef-5f39-4a95-969e-b0872e6db0d6",
		Name:         "Knight of the Keep",
		Completeness: CompletenessFull,
	})
}
