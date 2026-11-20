package effects

// Gold-Forged Sentinel — Artifact Creature — Chimera {6}, 4/4 (Journey into Nyx, #5558):
//
//	"Flying"
//
// Ornithopter's printed keyword.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "c338fb29-2dd2-4f76-ba3d-fa6770adcce8",
		Name:            "Gold-Forged Sentinel",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying"},
	})
}
