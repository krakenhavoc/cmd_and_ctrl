package effects

// Ripscale Predator — Creature — Dinosaur {4}{R}{R}, 6/5 (Ravnica: Clue Edition, #5929):
//
//	"Menace"
//
// Menace is a printed keyword.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "4c297085-5ec1-4d29-9a8e-1785c016ce09",
		Name:            "Ripscale Predator",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"menace"},
	})
}
