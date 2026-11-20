package effects

// Angelic Wall — Creature — Wall {1}{W}, 0/4 (Portal Second Age,
// #5625):
//
//	"Defender (This creature can't attack.)
//	 Flying"
//
// Defender and flying are engine keywords carried as printed data: it
// can't attack (CR 702.3b), and as a flier it can block a creature with
// flying.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "4502b24f-604b-4e36-9168-31c1a1ab4dab",
		Name:            "Angelic Wall",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"defender", "flying"},
	})
}
