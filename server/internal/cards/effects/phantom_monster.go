package effects

// Phantom Monster — Creature — Illusion {3}{U}, 3/3
// (Limited Edition Alpha, #5642): "Flying". A printed keyword and
// nothing else.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "bbd30183-524c-4b93-b953-90853ec3f39f",
		Name:            "Phantom Monster",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying"},
	})
}
