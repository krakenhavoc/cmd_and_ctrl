package effects

// Lightning Angel — Creature — Angel {1}{U}{R}{W}, 3/4 (Apocalypse,
// #5614):
//
//	"Flying, vigilance, haste"
//
// Three printed keywords on the PrintedKeywords slot (Serra Angel's
// shape), so the engine's own flying, vigilance and haste rules apply.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "2be85439-4606-4e1e-8be7-56d9e31f77c4",
		Name:            "Lightning Angel",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying", "vigilance", "haste"},
	})
}
