package effects

// Hunding Gjornersen — Legendary Creature — Human Warrior {3}{W}{U}{U}, 5/4 (Legends,
// #5769):
//
//	"Rampage 1 (Whenever this creature becomes blocked, it gets +1/+1
//	 until end of turn for each creature blocking it beyond the first.)"
//
// Rampage is the engine's keyword trigger (CR 702.23, game/rampage.go).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "dd4f3a78-3167-42d0-8334-5ac36750a253",
		Name:            "Hunding Gjornersen",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"rampage 1"},
	})
}
