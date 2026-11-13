package effects

// Frost Giant — Creature — Giant {3}{R}{R}{R}, 4/4 (Legends, #5769):
//
//	"Rampage 2 (Whenever this creature becomes blocked, it gets +2/+2
//	 until end of turn for each creature blocking it beyond the first.)"
//
// Rampage is the engine's keyword trigger (CR 702.23, game/rampage.go).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "56ed6af3-8f85-4907-9364-eca71f21bb11",
		Name:            "Frost Giant",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"rampage 2"},
	})
}
