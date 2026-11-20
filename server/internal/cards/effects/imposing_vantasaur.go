package effects

// Imposing Vantasaur — Creature — Dinosaur {5}{W}, 3/6 (Ikoria: Lair of Behemoths, #5512):
//
//	"Vigilance
//	 Cycling {1} ({1}, Discard this card: Draw a card.)"
//
// The printed keyword, and cycling from hand (CR 702.29a).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "a0d5fce0-a439-496a-9fcf-9271503f3ce8",
		Name:            "Imposing Vantasaur",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"vigilance"},
		Activated:       []ActivatedAbility{Cycling("{1}")},
	})
}
