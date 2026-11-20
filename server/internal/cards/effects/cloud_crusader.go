package effects

// Cloud Crusader — Creature — Human Knight {2}{W}{W}, 2/3 (Magic 2011,
// #5579):
//
//	"Flying
//	 First strike"
//
// Both are printed engine keywords.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "2ef481fa-ff20-426a-af16-fbf761ea5eaf",
		Name:            "Cloud Crusader",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying", "first strike"},
	})
}
