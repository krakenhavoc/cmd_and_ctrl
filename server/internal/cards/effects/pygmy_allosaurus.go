package effects

// Pygmy Allosaurus — Creature — Dinosaur {2}{G}, 2/2 (Ice Age, #5636):
//
//	"Swampwalk (This creature can't be blocked as long as defending
//	 player controls a Swamp.)"
//
// Swampwalk is a printed keyword the block legality check reads
// (CR 702.14c): any land with the Swamp subtype counts, basic or not.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "8c4b1ab2-01d9-486a-a9db-ac7c72500d30",
		Name:            "Pygmy Allosaurus",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"swampwalk"},
	})
}
