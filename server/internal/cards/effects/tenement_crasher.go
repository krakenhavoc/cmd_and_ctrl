package effects

// Tenement Crasher — Creature — Beast {5}{R}, 5/4 (Return to Ravnica, #5566):
//
//	"Haste"
//
// The printed keywords ride PrintedKeywords; the engine enforces them.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "ee11cecc-4772-4bb2-9af8-a2a99eb037f9",
		Name:            "Tenement Crasher",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"haste"},
	})
}
