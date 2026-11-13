package effects

// Tormented Angel — Creature — Angel {3}{W}, 1/5 (Urza's Destiny, #5621):
//
//	"Flying"
//
// A vanilla 1/5 flier: flying rides PrintedKeywords and the block
// check reads it.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "176eeafd-0482-4fa4-9174-10f8b9494d9b",
		Name:            "Tormented Angel",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying"},
	})
}
