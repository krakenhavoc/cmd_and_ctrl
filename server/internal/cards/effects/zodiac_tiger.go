package effects

// Zodiac Tiger — Creature — Cat {2}{G}{G}, 3/4
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
		OracleID:        "79aba9a7-b957-4139-8d50-f1c91fa17793",
		Name:            "Zodiac Tiger",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"forestwalk"},
	})
}
