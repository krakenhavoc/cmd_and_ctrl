package effects

// Battle Hurda — Creature — Giant {4}{W}, 3/3 (Worldwake, #5582):
//
//	"First strike"
//
// The printed keywords are engine keywords carried as data, as on
// Benalish Knight.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "46762944-8364-45c4-8008-e693e97ddc46",
		Name:            "Battle Hurda",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"first strike"},
	})
}
