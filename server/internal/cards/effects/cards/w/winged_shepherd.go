package cardsw

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Winged Shepherd — Creature — Angel {5}{W}, 3/3
// (Amonkhet, #5535):
//
//	"Flying, vigilance
//	 Cycling {W} ({W}, Discard this card: Draw a card.)"
//
// Cycling is the shared constructor; the cycle draws through the ordinary
// activation from hand.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "e18c93bc-2073-4bb6-83b4-c5c70f2d8a82",
		Name:            "Winged Shepherd",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying", "vigilance"},
		Activated:       []effects.ActivatedAbility{effects.Cycling("{W}")},
	})
}
