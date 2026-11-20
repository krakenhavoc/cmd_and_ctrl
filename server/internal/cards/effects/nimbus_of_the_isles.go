package effects

// Nimbus of the Isles — Creature — Elemental {4}{U}, 3/3 (Magic 2015,
// #5556):
//
//	"Flying"
//
// Armored Pegasus's printed flying, carried as data.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "3d529824-8e82-4c44-8e28-ed3419740c11",
		Name:            "Nimbus of the Isles",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying"},
	})
}
