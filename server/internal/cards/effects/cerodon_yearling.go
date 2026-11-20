package effects

// Cerodon Yearling — Creature — Beast {R}{W}, 2/2 (Alara Reborn, #5586):
//
//	"Vigilance, haste"
//
// Two printed keywords the engine enforces.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "6db09a0a-6c8e-4829-9b55-00745f0587b5",
		Name:            "Cerodon Yearling",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"vigilance", "haste"},
	})
}
