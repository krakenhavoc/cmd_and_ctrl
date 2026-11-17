package effects

// Marsh Goblins — Creature — Goblin {B}{R}, 1/1
// (The Dark, #5638): "Swampwalk". A printed keyword and nothing else.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "a2c5a504-5c8f-4c1b-a8b1-4e045aee4673",
		Name:            "Marsh Goblins",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"swampwalk"},
	})
}
