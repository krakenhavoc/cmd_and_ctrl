package effects

// Flying Men — Creature — Human {U}, 1/1 (Arabian Nights, #5641):
//
//	"Flying"
//
// Flying is an engine keyword, carried as printed data.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "b5666c68-059d-4e32-9b04-548b3058430d",
		Name:            "Flying Men",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying"},
	})
}
