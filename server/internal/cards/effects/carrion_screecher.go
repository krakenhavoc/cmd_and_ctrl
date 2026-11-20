package effects

// Carrion Screecher — Creature — Zombie Bird {3}{B}, 3/1
// (Hour of Devastation, #5534):
//
//	"Flying"
//
// Every word is an engine keyword carried as printed data.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "7b813e46-34f6-4918-97a7-a4b2a2464e86",
		Name:            "Carrion Screecher",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying"},
	})
}
