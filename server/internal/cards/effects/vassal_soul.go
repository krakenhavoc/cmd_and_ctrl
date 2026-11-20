package effects

// Vassal Soul — Creature — Spirit {1}{W/U}{W/U}, 2/2 (Return to Ravnica, #5566):
//
//	"Flying"
//
// The printed keywords ride PrintedKeywords; the engine enforces them.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "ae2329a1-e37b-44aa-8766-077910fa1197",
		Name:            "Vassal Soul",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying"},
	})
}
