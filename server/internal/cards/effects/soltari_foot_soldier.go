package effects

// Soltari Foot Soldier — Creature — Soltari Soldier {W}, 1/1
// (Tempest, #5628):
//
//	"Shadow (This creature can block or be blocked by only creatures
//	 with shadow.)"
//
// One printed keyword and nothing else. Shadow rides PrintedKeywords, and
// the block rules already allow a block between two creatures only when
// both or neither have shadow (CR 702.28b).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "6ccb3ac8-0509-431b-a196-d3cbf4fa80ae",
		Name:            "Soltari Foot Soldier",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"shadow"},
	})
}
