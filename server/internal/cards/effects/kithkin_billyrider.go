package effects

// Kithkin Billyrider — Creature — Kithkin Knight 1/3 {2}{W} (March of the Machine, S82):
//
//	"Double strike"
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "142d2063-d692-42a6-b828-42d998914589",
		Name:            "Kithkin Billyrider",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"double strike"},
	})
}
