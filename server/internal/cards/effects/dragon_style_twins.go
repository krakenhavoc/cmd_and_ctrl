package effects

// Dragon-Style Twins — Creature — Human Monk {3}{R}{R}, 3/3 (Khans of
// Tarkir, #5554):
//
//	"Double strike
//	 Prowess (Whenever you cast a noncreature spell, this creature gets
//	 +1/+1 until end of turn.)"
//
// Double strike is an engine keyword (CR 702.4b). Prowess is derived by
// the engine from the keyword (game/prowess.go).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "d150f456-c33a-4b19-ac73-b1d818d5db8f",
		Name:            "Dragon-Style Twins",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"double strike", "prowess"},
	})
}
