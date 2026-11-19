package effects

// Rakdos Shred-Freak — Creature — Human Berserker {B/R}{B/R}, 2/1 (Ravnica: Clue Edition, #5929):
//
//	"Haste"
//
// Haste is a printed keyword; the hybrid cost comes from the oracle data.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "e900b2b5-4bb2-482d-89ee-b5bbaf246698",
		Name:            "Rakdos Shred-Freak",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"haste"},
	})
}
