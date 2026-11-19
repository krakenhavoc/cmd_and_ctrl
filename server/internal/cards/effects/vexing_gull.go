package effects

// Vexing Gull — Creature — Bird {2}{U}, 2/2 (Foundations Jumpstart, S82,
// #5929):
//
//	"Flash
//	 Flying"
//
// Two printed keywords and nothing else.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "1a36d2a8-6117-4828-bbe2-bc63212d29d8",
		Name:            "Vexing Gull",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flash", "flying"},
	})
}
