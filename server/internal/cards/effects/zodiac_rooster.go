package effects

// Zodiac Rooster — Creature — Bird {1}{G}, 2/1
// (Portal Three Kingdoms, #5622):
//
//	"Plainswalk (This creature can't be blocked as long as defending player controls a Plains.)"
//
// Plainswalk is an engine keyword carried as printed data: it can't be
// blocked while the defending player controls a Plains (CR 702.14c).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "7280b7c5-e4d0-4a20-8634-5e240e6b3ca0",
		Name:            "Zodiac Rooster",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"plainswalk"},
	})
}
