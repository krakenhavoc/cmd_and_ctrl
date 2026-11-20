package effects

// Moaning Wall — Creature — Zombie Wall {2}{B}, 0/5
// (Hour of Devastation, #5534):
//
//	"Defender
//	 Cycling {2} ({2}, Discard this card: Draw a card.)"
//
// The keywords are printed data; cycling is the shared Cycling row, a
// hand ability that discards the card as its cost and draws (CR 702.29a).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "7bf22390-a6c0-4da6-9f05-e66628f72079",
		Name:            "Moaning Wall",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"defender"},
		Activated:       []ActivatedAbility{Cycling("{2}")},
	})
}
