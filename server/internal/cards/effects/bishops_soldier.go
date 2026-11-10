package effects

// Bishop's Soldier — Creature — Vampire Soldier {1}{W}, 2/2 (Foundations, #4465):
//
//	"Lifelink"
//
// Lifelink is an engine keyword, carried as printed data.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "84487586-4f71-457a-b983-4f1f7a71deef",
		Name:            "Bishop's Soldier",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"lifelink"},
	})
}
