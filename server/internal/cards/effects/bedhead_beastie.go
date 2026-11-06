package effects

// Bedhead Beastie — Creature — Beast {4}{R}{R}, 5/6 (Duskmourn: House of
// Horror, #4466):
//
//	"Menace (This creature can't be blocked except by two or more
//	 creatures.)
//	 Mountaincycling {2} ({2}, Discard this card: Search your library
//	 for a Mountain card, reveal it, put it into your hand, then
//	 shuffle.)"
//
// Menace is the printed keyword. Mountaincycling is Oliphaunt's
// Typecycling row: "a Mountain card" is any land card with the Mountain
// type, basic or not.
//
// No simplification.
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
