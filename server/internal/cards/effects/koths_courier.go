package effects

// Koth's Courier — Creature — Human Rogue {1}{R}{R}, 2/3 (Mirrodin Besieged,
// #5576):
//
//	"Forestwalk (This creature can't be blocked as long as defending
//	 player controls a Forest.)"
//
// Forestwalk is enforced at the block declaration: it can't be blocked
// while the defending player controls a Forest (CR 702.14c).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "59f57c25-ebb7-49df-8cdd-c6f069d1358d",
		Name:            "Koth's Courier",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"forestwalk"},
	})
}
