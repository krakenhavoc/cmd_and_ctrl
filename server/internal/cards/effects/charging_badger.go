package effects

// Charging Badger — Creature — Badger {G}, 1/1 (Born of the Gods,
// #5559):
//
//	"Trample"
//
// A keyword body.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "04be1e69-2d5c-4931-8906-a6807b5b08ca",
		Name:            "Charging Badger",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"trample"},
	})
}
