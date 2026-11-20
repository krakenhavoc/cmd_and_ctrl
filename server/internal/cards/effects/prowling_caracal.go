package effects

// Prowling Caracal — Creature — Cat {1}{W}, 3/1 (Ravnica Allegiance, #5521):
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
		OracleID:     "0eabe5e9-cb03-4481-9360-5e79eccd0631",
		Name:         "Prowling Caracal",
		Completeness: CompletenessFull,
	})
}
