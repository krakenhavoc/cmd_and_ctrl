package effects

// Lumengrid Warden — Creature — Human Wizard {1}{U}, 1/3 (Mirrodin, #5606):
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
		OracleID:     "ee16a8f3-2522-4c36-b4de-65484dfb7455",
		Name:         "Lumengrid Warden",
		Completeness: CompletenessFull,
	})
}
