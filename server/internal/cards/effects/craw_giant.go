package effects

// Craw Giant — Creature — Giant {3}{G}{G}{G}{G}, 6/4 (Legends, #5769):
//
//	"Trample
//	 Rampage 2 (Whenever this creature becomes blocked, it gets
//	 +2/+2 until end of turn for each creature blocking it beyond the
//	 first.)"
//
// Trample is the engine's keyword and rampage its keyword trigger (CR
// 702.23, game/rampage.go).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "152b621e-4c18-438c-9999-77e2685fda79",
		Name:            "Craw Giant",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"trample", "rampage 2"},
	})
}
