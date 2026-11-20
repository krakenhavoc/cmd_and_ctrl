package effects

// Stalker Hag — Creature — Hag {B/G}{B/G}{B/G}, 3/2 (Eventide, #5590):
//
//	"Swampwalk, forestwalk (This creature can't be blocked as long as
//	 defending player controls a Swamp or a Forest.)"
//
// Two landwalk keywords (CR 702.14), Odylic Wraith's and Cat Warriors'.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "c6aa6091-5503-4c22-91af-f999112b0081",
		Name:            "Stalker Hag",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"swampwalk", "forestwalk"},
	})
}
