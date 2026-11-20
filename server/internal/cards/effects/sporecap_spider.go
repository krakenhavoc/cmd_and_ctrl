package effects

// Sporecap Spider — Creature — Spider {2}{G}, 1/5 (Rise of the Eldrazi,
// #5581):
//
//	"Reach"
//
// A keyword-only body (see giant_spider.go).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "e9b5d4f3-b98a-463a-b628-289c2df81204",
		Name:            "Sporecap Spider",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"reach"},
	})
}
