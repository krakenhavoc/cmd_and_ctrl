package effects

// Ukud Cobra — Creature — Snake {3}{B}, 2/5 (Foundations Jumpstart):
//
//	"Deathtouch"
//
// Deathtouch is an engine keyword.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "d38cbb37-0ddd-4768-bb05-f35b38730e44",
		Name:            "Ukud Cobra",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"deathtouch"},
	})
}
