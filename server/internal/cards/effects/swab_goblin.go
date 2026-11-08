package effects

// Swab Goblin — Creature — Goblin Pirate {1}{R}, 2/2 (Foundations,
// #4465):
//
//	(no rules text)
//
// A vanilla body, registered so the catalog knows it (see
// fire_elemental.go). The cost, types and body ride game.Card from
// Scryfall, so the Spec is the key and the name.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "343ecf2f-0bbc-4840-8fb3-f5f61b26e76f",
		Name:         "Swab Goblin",
		Completeness: CompletenessFull,
	})
}
