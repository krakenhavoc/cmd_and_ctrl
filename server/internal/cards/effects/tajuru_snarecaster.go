package effects

// Tajuru Snarecaster — Creature — Elf Rogue {2}{G}, 1/4 (Zendikar Rising,
// #5508):
//
//	"Reach"
//
// An engine keyword carried as printed data (CR 702.17b).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "2061fe08-0c7c-4d79-a144-996c3012952c",
		Name:            "Tajuru Snarecaster",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"reach"},
	})
}
