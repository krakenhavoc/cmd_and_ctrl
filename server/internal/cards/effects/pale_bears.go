package effects

// Pale Bears — Creature — Bear {2}{G}, 2/2 (Ice Age, #5636):
//
//	"Islandwalk (This creature can't be blocked as long as defending
//	 player controls an Island.)"
//
// Islandwalk is a printed keyword the block legality check reads
// (CR 702.14c), with the defender's land types read after layer 4.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "0acb070b-6e81-4e0e-925d-14da9b931355",
		Name:            "Pale Bears",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"islandwalk"},
	})
}
