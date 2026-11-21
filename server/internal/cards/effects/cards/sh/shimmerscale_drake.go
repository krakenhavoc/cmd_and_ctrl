package cardssh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Shimmerscale Drake — Creature — Drake {4}{U}, 3/4
// (Amonkhet, #5535):
//
//	"Flying
//	 Cycling {2} ({2}, Discard this card: Draw a card.)"
//
// Cycling is the shared constructor; the cycle draws through the ordinary
// activation from hand.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "f3bd5ca1-9854-4bc0-b108-b555808e1d5f",
		Name:            "Shimmerscale Drake",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Activated:       []effects.ActivatedAbility{effects.Cycling("{2}")},
	})
}
