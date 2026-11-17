package effects

// Soaring Drake — Creature — Drake {2}{U}, 2/3 (Dominaria United, S82):
//
//	"Flying"
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "6bc14f1e-1df8-4e72-b05a-a848faa3b080",
		Name:            "Soaring Drake",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying"},
	})
}
