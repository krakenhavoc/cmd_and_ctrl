package cardssy

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Tah-Crop Skirmisher — Creature — Snake Warrior {1}{U}, 2/1
// (Amonkhet, #5535):
//
//	"Embalm {3}{U} ({3}{U}, Exile this card from your graveyard: Create a token that's a copy of it, except it's a white Zombie Snake Warrior with no mana cost. Embalm only as a sorcery.)"
//
// Embalm is the shared constructor (embalm.go, Sacred Cat): the cost
// exiles the card from the graveyard and the token is a white Zombie copy
// that keeps its other types and its printed keywords.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "6fab0fa0-d082-414e-9bcf-edc0fc1f6211",
		Name:         "Tah-Crop Skirmisher",
		Completeness: effects.CompletenessFull,
		Activated:    []effects.ActivatedAbility{effects.Embalm("{3}{U}")},
	})
}
