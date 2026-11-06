package effects

// Fire Nation Ambushers — Creature — Human Soldier {2}{B}, 3/2 (Avatar: The Last
// Airbender Eternal, #5075):
//
//	"Flash"
//
// Flash is an engine keyword, carried as printed data.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "8ed27504-f86b-4ce5-9395-7029825176ab",
		Name:            "Fire Nation Ambushers",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flash"},
	})
}
