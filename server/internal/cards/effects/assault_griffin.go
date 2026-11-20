package effects

// Assault Griffin — Creature — Griffin {3}{W}, 3/2 (Magic 2011, #5579):
//
//	"Flying"
//
// Flying is the printed keyword the block check reads.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "802b9333-c256-410a-a635-a22bfeda679f",
		Name:            "Assault Griffin",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying"},
	})
}
