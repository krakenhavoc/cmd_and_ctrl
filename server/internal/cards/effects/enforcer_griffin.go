package effects

// Enforcer Griffin — Creature — Griffin {4}{W}, 3/4 (Ravnica: Clue Edition):
//
//	"Flying"
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "4e1f4211-5865-492d-b850-0d36ac12e857",
		Name:            "Enforcer Griffin",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying"},
	})
}
