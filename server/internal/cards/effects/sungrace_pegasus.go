package effects

// Sungrace Pegasus — {1}{W} Creature — Pegasus 1/2 (Foundations Jumpstart):
//
//	"Flying
//	 Lifelink"
//
// Both keywords ride PrintedKeywords. No simplification.
func init() {
	Register(Spec{
		OracleID:        "67f59250-d46a-4637-9493-165d23278817",
		Name:            "Sungrace Pegasus",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying", "lifelink"},
	})
}
