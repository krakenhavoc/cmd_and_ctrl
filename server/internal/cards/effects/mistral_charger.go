package effects

// Mistral Charger — Creature — Pegasus {1}{W}, 2/1 (Ravnica Remastered):
//
//	"Flying"
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "a084b37a-bfb2-4dde-9faa-4c55ad894300",
		Name:            "Mistral Charger",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying"},
	})
}
