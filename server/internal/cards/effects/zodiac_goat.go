package effects

// Zodiac Goat — Creature — Goat {R}, 1/1 (Portal Three Kingdoms, #5622):
//
//	"Mountainwalk (This creature can't be blocked as long as defending player controls a Mountain.)"
//
// Mountainwalk is an engine keyword carried as printed data: it can't be
// blocked while the defending player controls a Mountain (CR 702.14c).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "01800a0c-ad93-4e45-a904-38253a143638",
		Name:            "Zodiac Goat",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"mountainwalk"},
	})
}
