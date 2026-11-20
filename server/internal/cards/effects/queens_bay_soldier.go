package effects

// Queen's Bay Soldier — Creature — Vampire Soldier {1}{B}, 2/2 (Ixalan,
// #5532):
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
		OracleID:     "b288c863-5377-473d-a628-bb199993a29c",
		Name:         "Queen's Bay Soldier",
		Completeness: CompletenessFull,
	})
}
