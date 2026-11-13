package effects

// Disciple of Grace — Creature — Human Cleric {1}{W}, 1/2 (Urza's Saga):
//
//	"Protection from black
//	 Cycling {2} ({2}, Discard this card: Draw a card.)"
//
// Protection is the printed keyword (CR 702.16a) and cycling is the
// shared Cycling row (CR 702.29a).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "7f4ceca9-0f55-4fa6-8e64-befbc4303d4e",
		Name:            "Disciple of Grace",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"protection from black"},
		Activated:       []ActivatedAbility{Cycling("{2}")},
	})
}
