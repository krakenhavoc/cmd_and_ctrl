package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Trueheart Duelist — Creature — Human Warrior {1}{W}, 2/2
// (Amonkhet, #5535):
//
//	"This creature can block an additional creature each combat.
//	 Embalm {2}{W} ({2}{W}, Exile this card from your graveyard: Create a token that's a copy of it, except it's a white Zombie Human Warrior with no mana cost. Embalm only as a sorcery.)"
//
// Embalm is the shared constructor (embalm.go, Sacred Cat): the cost
// exiles the card from the graveyard and the token is a white Zombie copy
// that keeps its other types and its printed keywords.
// The extra block is Foriysian Brigade's static.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "a0185db5-8988-4f63-8f3c-a19b4807a431",
		Name:         "Trueheart Duelist",
		Completeness: CompletenessFull,
		Static:       []game.StaticAbility{CanBlockAdditional(selfOnly, 1)},
		Activated:    []ActivatedAbility{Embalm("{2}{W}")},
	})
}
