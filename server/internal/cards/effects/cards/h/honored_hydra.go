package cardsh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Honored Hydra — Creature — Snake Hydra {5}{G}, 6/6
// (Amonkhet, #5535):
//
//	"Trample
//	 Embalm {3}{G} ({3}{G}, Exile this card from your graveyard: Create a token that's a copy of it, except it's a white Zombie Snake Hydra with no mana cost. Embalm only as a sorcery.)"
//
// Embalm is the shared constructor (embalm.go, Sacred Cat): the cost
// exiles the card from the graveyard and the token is a white Zombie copy
// that keeps its other types and its printed keywords.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "db63d999-97a6-4a62-83bd-e0677e6f8956",
		Name:            "Honored Hydra",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"trample"},
		Activated:       []effects.ActivatedAbility{effects.Embalm("{3}{G}")},
	})
}
