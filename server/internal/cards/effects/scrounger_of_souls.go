package effects

// Scrounger of Souls — Creature — Horror {4}{B}, 3/4
// (Hour of Devastation, #5534):
//
//	"Lifelink"
//
// Every word is an engine keyword carried as printed data.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "23dd6d4a-d933-4328-8d09-75f7908ae347",
		Name:            "Scrounger of Souls",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"lifelink"},
	})
}
