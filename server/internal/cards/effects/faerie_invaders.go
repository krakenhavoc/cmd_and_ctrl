package effects

// Faerie Invaders — Creature — Faerie Rogue {4}{U}, 3/3 (Magic 2013, #5568):
//
//	"Flash (You may cast this spell any time you could cast an instant.)
//	 Flying"
//
// A keyword body: flash (CR 702.8) lets it be cast at instant speed, and
// flying (CR 702.9) restricts its blockers.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "df7580fb-4c37-4c09-8785-e320cb124621",
		Name:            "Faerie Invaders",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flash", "flying"},
	})
}
