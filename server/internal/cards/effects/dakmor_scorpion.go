package effects

// Dakmor Scorpion — Creature — Scorpion {1}{B}, 2/1 (Portal Second Age,
// #5625):
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
		OracleID:     "008290c2-499d-425b-9a6c-c4aeb47c33ff",
		Name:         "Dakmor Scorpion",
		Completeness: CompletenessFull,
	})
}
