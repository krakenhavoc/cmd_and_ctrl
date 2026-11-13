package effects

// Moor Fiend — Creature — Horror {3}{B}, 3/3 (Ice Age, #5636):
//
//	"Swampwalk (This creature can't be blocked as long as defending
//	 player controls a Swamp.)"
//
// Marsh Boa's swampwalk: a printed keyword the block legality check
// reads (CR 702.14c), with the defender's land types read after
// layer 4.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "ff1fd718-ee97-425c-b8bf-97fea192c631",
		Name:            "Moor Fiend",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"swampwalk"},
	})
}
