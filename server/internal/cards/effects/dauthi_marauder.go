package effects

// Dauthi Marauder — Creature — Dauthi Minion {2}{B}, 3/1 (Tempest, #5628):
//
//	"Shadow (This creature can block or be blocked by only creatures
//	 with shadow.)"
//
// Soltari Foot Soldier's shape: shadow is a printed keyword, and the
// block rules enforce it (CR 702.28b).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "cb754749-a769-4c9c-b62b-4972dfc5069f",
		Name:            "Dauthi Marauder",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"shadow"},
	})
}
