package effects

// Glissa's Courier — Creature — Phyrexian Horror {1}{G}{G}, 2/3 (Mirrodin Besieged,
// #5576):
//
//	"Mountainwalk (This creature can't be blocked as long as defending
//	 player controls a Mountain.)"
//
// Mountainwalk is enforced at the block declaration: it can't be blocked
// while the defending player controls a Mountain (CR 702.14c).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "25fb2a67-fd7f-4a06-bd59-118d82c86d6e",
		Name:            "Glissa's Courier",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"mountainwalk"},
	})
}
