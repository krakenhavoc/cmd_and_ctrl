package effects

// Angel of the God-Pharaoh — Creature — Angel {4}{W}{W}, 4/4
// (Hour of Devastation, #5534):
//
//	"Flying
//	 Cycling {2} ({2}, Discard this card: Draw a card.)"
//
// The keywords are printed data; cycling is the shared Cycling row, a
// hand ability that discards the card as its cost and draws (CR 702.29a).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "701912d5-9995-4ac7-9858-188f78b2b8a3",
		Name:            "Angel of the God-Pharaoh",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Activated:       []ActivatedAbility{Cycling("{2}")},
	})
}
