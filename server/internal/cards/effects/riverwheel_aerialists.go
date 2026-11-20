package effects

// Riverwheel Aerialists — Creature — Djinn Monk {5}{U}, 4/5 (Khans of
// Tarkir, #5554):
//
//	"Flying
//	 Prowess (Whenever you cast a noncreature spell, this creature gets
//	 +1/+1 until end of turn.)"
//
// Flying is printed data (CR 702.9b). Prowess is derived by the engine
// from the keyword (game/prowess.go).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "f9221367-24a5-41a3-8943-c6a3cb5d605d",
		Name:            "Riverwheel Aerialists",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying", "prowess"},
	})
}
