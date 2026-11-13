package effects

// Darkwatch Elves — Creature — Elf {2}{G}, 2/2 (Urza's Legacy, #5623):
//
//	"Protection from black
//	 Cycling {2} ({2}, Discard this card: Draw a card.)"
//
// Disciple of Grace's text on a green 2/2: protection is the printed keyword
// (CR 702.16a) and cycling is the shared Cycling row (CR 702.29a).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "55ad9924-f217-4844-8a82-48746186bc53",
		Name:            "Darkwatch Elves",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"protection from black"},
		Activated:       []ActivatedAbility{Cycling("{2}")},
	})
}
