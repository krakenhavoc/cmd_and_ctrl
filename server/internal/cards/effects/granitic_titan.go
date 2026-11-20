package effects

// Granitic Titan — Creature — Elemental {4}{R}{R}, 5/4
// (Hour of Devastation, #5534):
//
//	"Menace
//	 Cycling {2} ({2}, Discard this card: Draw a card.)"
//
// The keywords are printed data; cycling is the shared Cycling row, a
// hand ability that discards the card as its cost and draws (CR 702.29a).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "d398868b-7574-40d5-8915-6393ee489067",
		Name:            "Granitic Titan",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"menace"},
		Activated:       []ActivatedAbility{Cycling("{2}")},
	})
}
