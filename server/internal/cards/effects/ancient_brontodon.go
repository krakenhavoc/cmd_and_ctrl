package effects

// Ancient Brontodon — Creature — Dinosaur {6}{G}{G}, 9/9 (Ixalan, #5532):
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
		OracleID:     "745be03c-69dc-4239-a5e7-38587b56eac0",
		Name:         "Ancient Brontodon",
		Completeness: CompletenessFull,
	})
}
