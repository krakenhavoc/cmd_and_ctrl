package effects

// Aven Initiate — Creature — Bird Warrior {3}{U}, 3/2
// (Amonkhet, #5535):
//
//	"Flying
//	 Embalm {6}{U} ({6}{U}, Exile this card from your graveyard: Create a token that's a copy of it, except it's a white Zombie Bird Warrior with no mana cost. Embalm only as a sorcery.)"
//
// Embalm is the shared constructor (embalm.go, Sacred Cat): the cost
// exiles the card from the graveyard and the token is a white Zombie copy
// that keeps its other types and its printed keywords.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "c872012d-b598-4076-8385-2e3771c0306f",
		Name:            "Aven Initiate",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Activated:       []ActivatedAbility{Embalm("{6}{U}")},
	})
}
