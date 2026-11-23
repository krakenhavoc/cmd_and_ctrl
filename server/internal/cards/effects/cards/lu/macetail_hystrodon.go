package cardslu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Macetail Hystrodon — Creature — Beast {6}{R}, 4/4 (Dominaria Remastered,
// #5948):
//
//	"First strike, haste
//	 Cycling {3} ({3}, Discard this card: Draw a card.)"
//
// Both keywords ride PrintedKeywords; cycling is the shared Cycling
// activated ability.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "2f1e2742-d7df-4893-abc8-cb927c500569",
		Name:            "Macetail Hystrodon",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"first strike", "haste"},
		Activated:       []effects.ActivatedAbility{effects.Cycling("{3}")},
	})
}
