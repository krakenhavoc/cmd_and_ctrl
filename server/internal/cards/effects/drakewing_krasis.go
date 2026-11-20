package effects

// Drakewing Krasis — Creature — Lizard Drake {1}{G}{U}, 3/1 (Gatecrash, #5565):
//
//	"Flying, trample"
//
// Flying and trample are engine keywords.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "782e8499-0714-4cee-869e-8ebd7dad0242",
		Name:            "Drakewing Krasis",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying", "trample"},
	})
}
