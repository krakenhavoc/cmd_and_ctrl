package effects

// Alaborn Musketeer — Creature — Human Soldier {1}{W}, 2/1 (Portal Second Age,
// #5625):
//
//	"Reach (This creature can block creatures with flying.)"
//
// Reach is an engine keyword carried as printed data: it can block a
// creature with flying (CR 702.17b).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "61daa88d-fca4-4019-97f3-568346157ba6",
		Name:            "Alaborn Musketeer",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"reach"},
	})
}
