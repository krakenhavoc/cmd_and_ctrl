package effects

// Trained Caracal — Creature — Cat {W}, 1/1 (Return to Ravnica, #5566):
//
//	"Lifelink (Damage dealt by this creature also causes you to gain
//	 that much life.)"
//
// The printed keywords ride PrintedKeywords; the engine enforces them.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "a939f01b-afce-4d92-a897-1a6dc3d788c6",
		Name:            "Trained Caracal",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"lifelink"},
	})
}
