package effects

// Wall of Light — Creature — Wall {2}{W}, 1/5 (Legends, #5639):
//
//	"Defender (This creature can't attack.)
//	 Protection from black"
//
// Defender keeps it from attacking (CR 702.3b), and protection from
// black refuses black spells' and black sources' targets (CR 702.16b).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "d0c05cad-6e14-4726-b7fc-8f2fc7df7267",
		Name:            "Wall of Light",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"defender", "protection from black"},
	})
}
