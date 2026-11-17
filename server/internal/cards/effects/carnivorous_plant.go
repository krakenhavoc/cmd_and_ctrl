package effects

// Carnivorous Plant — Creature — Plant Wall {3}{G}, 4/5
// (The Dark, #5638): "Defender". A printed keyword and nothing else.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "3a7478d7-b33e-4227-8cc3-b30643624d56",
		Name:            "Carnivorous Plant",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"defender"},
	})
}
