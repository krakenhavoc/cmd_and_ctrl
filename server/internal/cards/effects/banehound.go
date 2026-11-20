package effects

// Banehound — Creature — Nightmare Dog {B}, 1/1 (War of the Spark, #5520):
//
//	"Lifelink, haste"
//
// The keywords are engine keywords carried as printed data; combat does
// the rest.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "63d846e2-4bef-4911-9508-273c75a4a040",
		Name:            "Banehound",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"lifelink", "haste"},
	})
}
