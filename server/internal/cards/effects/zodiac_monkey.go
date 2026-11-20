package effects

// Zodiac Monkey — Creature — Monkey {1}{G}, 2/1
// (Portal Three Kingdoms, #5622):
//
//	"Forestwalk (This creature can't be blocked as long as defending player controls a Forest.)"
//
// Forestwalk is an engine keyword carried as printed data: it can't be
// blocked while the defending player controls a Forest (CR 702.14c).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "47dd7ef4-7645-4878-ab91-261aa8fcf834",
		Name:            "Zodiac Monkey",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"forestwalk"},
	})
}
