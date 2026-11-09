package effects

// Iron Giant — Artifact Creature — Demon {7}, 6/6 (Final Fantasy, #4462):
//
//	"Reach, vigilance, trample"
//
// Three printed combat keywords and nothing else; the engine reads them
// through the layers.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "325b4288-fe1b-4e29-ae13-72924b4db3dd",
		Name:            "Iron Giant",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"reach", "vigilance", "trample"},
	})
}
