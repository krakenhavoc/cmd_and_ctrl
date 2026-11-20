package effects

// Winged Shepherd — Creature — Angel {5}{W}, 3/3
// (Amonkhet, #5535):
//
//	"Flying, vigilance
//	 Cycling {W} ({W}, Discard this card: Draw a card.)"
//
// Cycling is the shared constructor; the cycle draws through the ordinary
// activation from hand.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "e18c93bc-2073-4bb6-83b4-c5c70f2d8a82",
		Name:            "Winged Shepherd",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying", "vigilance"},
		Activated:       []ActivatedAbility{Cycling("{W}")},
	})
}
