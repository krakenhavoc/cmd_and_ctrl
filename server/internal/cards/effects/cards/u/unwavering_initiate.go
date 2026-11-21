package cardsu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Unwavering Initiate — Creature — Human Warrior {2}{W}, 3/2
// (Amonkhet, #5535):
//
//	"Vigilance
//	 Embalm {4}{W} ({4}{W}, Exile this card from your graveyard: Create a token that's a copy of it, except it's a white Zombie Human Warrior with no mana cost. Embalm only as a sorcery.)"
//
// Embalm is the shared constructor (embalm.go, Sacred Cat): the cost
// exiles the card from the graveyard and the token is a white Zombie copy
// that keeps its other types and its printed keywords.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "e6088c23-a660-4d2c-b633-1c6cb1a59ebe",
		Name:            "Unwavering Initiate",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"vigilance"},
		Activated:       []effects.ActivatedAbility{effects.Embalm("{4}{W}")},
	})
}
