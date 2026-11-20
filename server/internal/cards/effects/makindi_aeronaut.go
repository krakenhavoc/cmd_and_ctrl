package effects

// Makindi Aeronaut — Creature — Kor Scout Ally {1}{W}, 1/3 (Oath of the Gatewatch, #5545):
//
//	"Flying"
//
// Its whole text is a printed keyword.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "9f980829-90c1-43e2-a103-915f38317ed0",
		Name:            "Makindi Aeronaut",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying"},
	})
}
