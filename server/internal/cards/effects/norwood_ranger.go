package effects

// Norwood Ranger — Creature — Elf Scout Ranger {G}, 1/2 (Portal Second Age,
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
		OracleID:     "e62ec170-c588-41ff-8be8-b05a125693a1",
		Name:         "Norwood Ranger",
		Completeness: CompletenessFull,
	})
}
