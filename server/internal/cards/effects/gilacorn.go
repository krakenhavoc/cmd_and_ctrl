package effects

// Gilacorn — Creature — Lizard {B}, 1/1 (Avatar: The Last
// Airbender Eternal, #5075):
//
//	"Deathtouch"
//
// Deathtouch is an engine keyword, carried as printed data.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "1551cd95-1e68-404f-be1d-cc9c7f447210",
		Name:            "Gilacorn",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"deathtouch"},
	})
}
