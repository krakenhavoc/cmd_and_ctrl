package effects

// Emerald Oryx — Creature — Antelope {3}{G}, 2/3 (Magic 2010, #5585):
//
//	"Forestwalk (This creature can't be blocked as long as defending
//	 player controls a Forest.)"
//
// Lynx's forestwalk (CR 702.14c), carried as printed data.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "4c26c796-26ec-41f7-b9ca-58f4eb935c47",
		Name:            "Emerald Oryx",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"forestwalk"},
	})
}
