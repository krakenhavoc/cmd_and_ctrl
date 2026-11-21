package effects

// Rorix Bladewing — Legendary Creature — Dragon {3}{R}{R}{R}, 6/5
// (Commander Masters):
//
//	"Flying, haste"
//
// Two printed keywords, nothing else.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "48fe0f52-1676-49d5-acf6-5476081e8e1b",
		Name:            "Rorix Bladewing",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying", "haste"},
	})
}
