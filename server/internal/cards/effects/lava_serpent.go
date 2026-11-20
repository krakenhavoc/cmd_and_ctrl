package effects

// Lava Serpent — Creature — Elemental Serpent {5}{R}, 5/5 (Ikoria: Lair of Behemoths, #5512):
//
//	"Haste
//	 Cycling {2} ({2}, Discard this card: Draw a card.)"
//
// The printed keyword, and cycling from hand (CR 702.29a).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "4d94d807-97d9-4cdf-a63c-d8b82c2acc6e",
		Name:            "Lava Serpent",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"haste"},
		Activated:       []ActivatedAbility{Cycling("{2}")},
	})
}
