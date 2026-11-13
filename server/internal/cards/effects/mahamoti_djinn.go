package effects

// Mahamoti Djinn — Creature — Djinn {4}{U}{U}, 5/6
// (Limited Edition Alpha, #5642): "Flying". A printed keyword and
// nothing else.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "c39ea5f9-6ec0-4697-897b-779e326754a7",
		Name:            "Mahamoti Djinn",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying"},
	})
}
