package effects

// Brazen Scourge — Creature — Gremlin {1}{R}{R}, 3/3 (Foundations, #4465):
//
//	"Haste"
//
// Haste is an engine keyword, carried as printed data.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "422ef4ec-863a-450e-9577-a0707bc16532",
		Name:            "Brazen Scourge",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"haste"},
	})
}
