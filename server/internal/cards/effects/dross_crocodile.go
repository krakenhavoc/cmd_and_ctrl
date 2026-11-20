package effects

// Dross Crocodile — Creature — Zombie Crocodile {3}{B}, 5/1 (Fifth Dawn, #5604):
//
//	(no rules text)
//
// A vanilla body, registered so the catalog knows it (see coral_eel.go).
// The cost, types and body ride game.Card from Scryfall.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "028fa430-722d-45aa-b415-3f84076091a9",
		Name:         "Dross Crocodile",
		Completeness: CompletenessFull,
	})
}
