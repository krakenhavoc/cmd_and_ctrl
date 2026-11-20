package effects

// Cursed Minotaur — Creature — Zombie Minotaur {2}{B}, 3/2
// (Amonkhet, #5535):
//
//	"Menace (This creature can't be blocked except by two or more creatures.)"
//
// A printed keyword the engine runs off the card; the Spec declares it.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "88b50194-bc39-47a6-968b-983553d89cf9",
		Name:            "Cursed Minotaur",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"menace"},
	})
}
