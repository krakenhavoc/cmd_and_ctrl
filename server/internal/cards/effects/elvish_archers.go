package effects

// Elvish Archers — Creature — Elf Archer {1}{G}, 2/1
// (Limited Edition Alpha, #5642): "First strike". A printed keyword and
// nothing else.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "534ce8e2-53a1-407c-9881-d3b5347d43c3",
		Name:            "Elvish Archers",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"first strike"},
	})
}
