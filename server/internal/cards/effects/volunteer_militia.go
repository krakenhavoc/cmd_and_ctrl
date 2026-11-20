package effects

// Volunteer Militia — Creature — Human Soldier {W}, 1/2 (Portal Second Age,
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
		OracleID:     "4f598459-b776-4ac1-bc69-8fde21674451",
		Name:         "Volunteer Militia",
		Completeness: CompletenessFull,
	})
}
