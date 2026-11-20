package effects

// Zodiac Dog — Creature — Dog {2}{R}, 2/2 (Portal Three Kingdoms, #5622):
//
//	"Mountainwalk (This creature can't be blocked as long as defending player controls a Mountain.)"
//
// Mountainwalk is an engine keyword carried as printed data: it can't be
// blocked while the defending player controls a Mountain (CR 702.14c).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "4431ec11-9bd7-4750-8bc7-378beba06201",
		Name:            "Zodiac Dog",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"mountainwalk"},
	})
}
