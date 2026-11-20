package effects

// Daggerdrome Imp — Creature — Imp {1}{B}, 1/1 (Return to Ravnica, #5566):
//
//	"Flying / Lifelink (Damage dealt by this creature also causes you
//	 to gain that much life.)"
//
// The printed keywords ride PrintedKeywords; the engine enforces them.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "894979ba-e4c7-4914-b0d3-133cbaed38c2",
		Name:            "Daggerdrome Imp",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying", "lifelink"},
	})
}
