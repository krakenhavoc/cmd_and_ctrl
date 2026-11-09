package effects

// Bedhead Beastie — Creature — Beast {4}{R}{R}, 5/6 (Duskmourn, #4466):
//
//	"Menace
//	 Mountaincycling {2} ({2}, Discard this card: Search your library
//	 for a Mountain card, reveal it, put it into your hand, then
//	 shuffle.)"
//
// Mountaincycling is typecycling (CR 702.29e) for the Mountain subtype.
func init() {
	Register(Spec{
		OracleID:        "7143d82a-931a-4918-bb14-37bc5b9a4d99",
		Name:            "Bedhead Beastie",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"menace"},
		Activated: []ActivatedAbility{
			Typecycling("Mountaincycling", "{2}", "a Mountain card", IsLandWithSubtype("Mountain")),
		},
	})
}
