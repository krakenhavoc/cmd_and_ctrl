package effects

// Vernadi Shieldmate — Creature — Human Soldier {1}{G/W}, 2/2 (Ravnica: Clue Edition):
//
//	"Vigilance"
//
// The keyword rides PrintedKeywords; there is no other text.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "03e04034-48b5-4f1f-9859-2b46dedf4d8b",
		Name:            "Vernadi Shieldmate",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"vigilance"},
	})
}
