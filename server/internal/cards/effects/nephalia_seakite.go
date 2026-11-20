package effects

// Nephalia Seakite — Creature — Bird {3}{U}, 2/3 (Dark Ascension, #5571):
//
//	"Flash (You may cast this spell any time you could cast an
//	 instant.)
//	 Flying"
//
// Flash is the printed keyword the cast-timing gate reads; flying is
// printed data, as on Wind Strider.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "379ef166-a29f-4715-8ce4-9d7d567e4e91",
		Name:            "Nephalia Seakite",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flash", "flying"},
	})
}
