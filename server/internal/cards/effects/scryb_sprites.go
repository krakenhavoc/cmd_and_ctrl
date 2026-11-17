package effects

// Scryb Sprites — Creature — Faerie {G}, 1/1
// (Limited Edition Alpha, #5642): "Flying". A printed keyword and
// nothing else.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "a1f20695-6f08-4d5c-9fba-b0018bee298e",
		Name:            "Scryb Sprites",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying"},
	})
}
