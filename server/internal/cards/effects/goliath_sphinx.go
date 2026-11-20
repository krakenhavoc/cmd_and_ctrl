package effects

// Goliath Sphinx — Creature — Sphinx {5}{U}{U}, 8/7 (Worldwake, #5582):
//
//	"Flying"
//
// The printed keywords are engine keywords carried as data, as on
// Cloud Djinn.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "73115500-4468-43b3-9151-13e3da9050c5",
		Name:            "Goliath Sphinx",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying"},
	})
}
