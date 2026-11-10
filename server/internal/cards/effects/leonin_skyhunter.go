package effects

// Leonin Skyhunter — Creature — Cat Knight {W}{W}, 2/2 (Foundations, #4465):
//
//	"Flying"
//
// Flying is an engine keyword, carried as printed data.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "517c7295-8ba2-47a6-a1c6-aee722f8ca88",
		Name:            "Leonin Skyhunter",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying"},
	})
}
