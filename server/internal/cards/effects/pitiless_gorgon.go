package effects

// Pitiless Gorgon — Creature — Gorgon {1}{B/G}{B/G}, 2/2 (Ravnica: Clue Edition, #5929):
//
//	"Deathtouch"
//
// Deathtouch is a printed keyword; the hybrid cost comes from the oracle data.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "4704ce78-b287-4038-abdd-ddb20116874d",
		Name:            "Pitiless Gorgon",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"deathtouch"},
	})
}
