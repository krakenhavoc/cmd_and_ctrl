package effects

// Zodiac Rabbit — Creature — Rabbit {G}, 1/1 (Portal Three Kingdoms, #5622):
//
//	"Forestwalk (This creature can't be blocked as long as defending player controls a Forest.)"
//
// Forestwalk is an engine keyword carried as printed data: it can't be
// blocked while the defending player controls a Forest (CR 702.14c).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "048cc868-1657-448d-9f2c-97973e73f477",
		Name:            "Zodiac Rabbit",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"forestwalk"},
	})
}
