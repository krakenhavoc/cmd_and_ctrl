package effects

// Ghirapur Osprey — Creature — Bird {2}{W}, 2/2 (Aether Revolt, #5536):
// "Flying". A printed keyword and nothing else.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "b4de9edd-4a5b-4858-8454-38381df8b9a2",
		Name:            "Ghirapur Osprey",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying"},
	})
}
