package effects

// Sworn Guardian — Creature — Merfolk Warrior {1}{U}, 1/3 (Rivals of Ixalan, #5531):
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
		OracleID:     "9f63ab29-b86e-4337-b9dd-2b35c8415f90",
		Name:         "Sworn Guardian",
		Completeness: CompletenessFull,
	})
}
