package effects

// Tel-Jilad Archers — Creature — Elf Archer {4}{G}, 2/4 (Mirrodin, #5606):
//
//	"Reach (This creature can block creatures with flying.)
//	 Protection from artifacts"
//
// Two printed keywords: reach (CR 702.17b) and protection from
// artifacts (CR 702.16a).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "939203ec-3de4-44d8-932e-63a32e5f2d47",
		Name:            "Tel-Jilad Archers",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"reach", "protection from artifacts"},
	})
}
