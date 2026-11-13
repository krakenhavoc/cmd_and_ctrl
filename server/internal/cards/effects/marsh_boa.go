package effects

// Marsh Boa — Creature — Snake {G}, 1/1 (Prophecy, #5617):
//
//	"Swampwalk (This creature can't be blocked as long as defending
//	 player controls a Swamp.)"
//
// Dirtwater Wraith's swampwalk without the pump: a printed keyword the
// block legality check reads (CR 702.14c), with the defender's land
// types read after layer 4.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "246f7446-036d-469c-9583-5699ebdcec6d",
		Name:            "Marsh Boa",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"swampwalk"},
	})
}
