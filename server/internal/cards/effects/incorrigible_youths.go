package effects

// Incorrigible Youths — Creature — Vampire {3}{R}{R}, 4/3 (Foundations
// Jumpstart, BNS batch 32):
//
//	"Haste
//	 Madness {2}{R} (If you discard this card, discard it into exile. When
//	 you do, cast it for its madness cost or put it into your graveyard.)"
//
// Anje's Ravager's Madness declaration; haste is a printed keyword.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "9c239027-c5cb-43c6-b4f3-93c269a1d853",
		Name:            "Incorrigible Youths",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"haste"},
		Madness:         "{2}{R}",
	})
}
