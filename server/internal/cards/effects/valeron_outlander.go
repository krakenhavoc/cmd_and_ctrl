package effects

// Valeron Outlander — Creature — Human Scout {G}{W}, 2/2 (Conflux,
// #5587):
//
//	"Protection from black"
//
// Protection from black is a printed keyword, read by targeting,
// blocking, damage and attaching (CR 702.16a).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "01d70534-7c26-4a54-9eca-d0ebbefc2159",
		Name:            "Valeron Outlander",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"protection from black"},
	})
}
