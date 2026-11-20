package effects

// Nacatl Outlander — Creature — Cat Scout {R}{G}, 2/2 (Conflux,
// #5587):
//
//	"Protection from blue"
//
// Protection from blue is a printed keyword, read by targeting,
// blocking, damage and attaching (CR 702.16a).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "333555d6-3fcc-4021-ab60-afc36d723877",
		Name:            "Nacatl Outlander",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"protection from blue"},
	})
}
