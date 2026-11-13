package effects

// Mountain Goat — Creature — Goat {R}, 1/1 (Ice Age, #5636):
//
//	"Mountainwalk (This creature can't be blocked as long as defending
//	 player controls a Mountain.)"
//
// Vug Lizard's mountainwalk: a printed keyword the block legality check
// reads (CR 702.14c), against the defending player's lands only.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "031cd561-d9b1-4a0e-bc38-bc09e80563c3",
		Name:            "Mountain Goat",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"mountainwalk"},
	})
}
