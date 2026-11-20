package effects

// Silverclaw Griffin — Creature — Griffin {3}{W}{W}, 3/2 (Dark Ascension, #5571):
//
//	"Flying, first strike"
//
// Both keywords are printed data, as on Thunder Spirit.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "5a7ecb5c-9b22-4d2c-b851-db92c1d3c745",
		Name:            "Silverclaw Griffin",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying", "first strike"},
	})
}
