package effects

// Seagraf Skaab — Creature — Zombie {1}{U}, 1/3
// (Shadows over Innistrad, #5543):
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
		OracleID:     "1c7dd66f-a062-49ac-aa7c-1bf5ec069d73",
		Name:         "Seagraf Skaab",
		Completeness: CompletenessFull,
	})
}
