package effects

// Riptide Turtle — Creature — Turtle {1}{U}, 0/5 (Theros Beyond Death, #5514):
//
//	"Flash
//	 Defender"
//
// Flash and defender are the printed keywords.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "cb7e076c-a695-4c79-9f20-498370c81380",
		Name:            "Riptide Turtle",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flash", "defender"},
	})
}
