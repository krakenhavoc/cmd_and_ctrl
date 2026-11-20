package effects

// Striped Riverwinder — Creature — Serpent {6}{U}, 5/5
// (Hour of Devastation, #5534):
//
//	"Hexproof
//	 Cycling {U} ({U}, Discard this card: Draw a card.)"
//
// The keywords are printed data; cycling is the shared Cycling row, a
// hand ability that discards the card as its cost and draws (CR 702.29a).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "725972de-4816-44cf-a2a5-4326c3bc1b4b",
		Name:            "Striped Riverwinder",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"hexproof"},
		Activated:       []ActivatedAbility{Cycling("{U}")},
	})
}
