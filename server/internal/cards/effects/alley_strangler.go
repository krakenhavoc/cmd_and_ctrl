package effects

// Alley Strangler — Creature — Aetherborn Rogue {2}{B}, 2/3 (Aether Revolt, #5536):
// "Menace". A printed keyword and nothing else.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "fb9af562-7384-4575-93bb-34f1ada23735",
		Name:            "Alley Strangler",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"menace"},
	})
}
