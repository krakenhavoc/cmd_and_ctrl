package effects

// Risen Sanctuary — Creature — Elemental {5}{G}{W}, 8/8 (Return to Ravnica, #5566):
//
//	"Vigilance"
//
// The printed keywords ride PrintedKeywords; the engine enforces them.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "a2198d36-4e8c-4d64-bda5-a95d85398155",
		Name:            "Risen Sanctuary",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"vigilance"},
	})
}
