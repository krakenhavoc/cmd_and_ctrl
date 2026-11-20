package effects

// Prized Griffin — Creature — Griffin {4}{W}, 3/4 (Throne of Eldraine, #5516):
//
//	"Flying"
//
// A printed keyword the engine enforces on its own.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "d383902c-63ba-4714-bfeb-377829950b0d",
		Name:            "Prized Griffin",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying"},
	})
}
