package effects

// Oketra's Attendant — Creature — Bird Soldier {3}{W}{W}, 3/3
// (Amonkhet, #5535):
//
//	"Flying
//	 Cycling {2} ({2}, Discard this card: Draw a card.)
//	 Embalm {3}{W}{W} ({3}{W}{W}, Exile this card from your graveyard: Create a token that's a copy of it, except it's a white Zombie Bird Soldier with no mana cost. Embalm only as a sorcery.)"
//
// Cycling from hand and embalm from the graveyard, the two shared
// constructors, in printed order.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "c45d1df1-92a7-4b36-8e27-1f88c3737bf9",
		Name:            "Oketra's Attendant",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Activated:       []ActivatedAbility{Cycling("{2}"), Embalm("{3}{W}{W}")},
	})
}
