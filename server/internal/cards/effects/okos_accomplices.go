package effects

// Oko's Accomplices — Creature — Faerie {2}{U}, 2/3 (Throne of Eldraine, #5516):
//
//	"Flying"
//
// A printed keyword the engine enforces on its own.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "93cd12da-a7d5-4d01-a3bc-77e441cb0947",
		Name:            "Oko's Accomplices",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying"},
	})
}
