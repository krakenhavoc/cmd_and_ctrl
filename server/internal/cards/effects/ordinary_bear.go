package effects

// Ordinary Bear — Creature — Bear {3}{G}, 4/5 (The Hobbit, HOB):
//
//	(no rules text)
//
// A vanilla body, registered so the catalog knows it (see
// yargle_glutton_of_urborg.go). The cost, types and body ride game.Card
// from Scryfall, so the Spec is the key and the name (#4454).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "89a2fab8-b074-42aa-a871-d9c5de2d0895",
		Name:         "Ordinary Bear",
		Completeness: CompletenessFull,
	})
}
