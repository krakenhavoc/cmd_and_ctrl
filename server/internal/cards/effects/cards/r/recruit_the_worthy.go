package cardsr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Recruit the Worthy — Instant {W} (Modern Horizons, #5519):
//
//	"Buyback {3} (You may pay an additional {3} as you cast this spell. If
//	 you do, put this card into your hand as it resolves.)
//	 Create a 1/1 white Soldier creature token."
//
// Anoint's buyback on a one-Soldier spell.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:      "e25e2a3f-4add-4744-b1f8-2e880998a51f",
		Name:          "Recruit the Worthy",
		Completeness:  effects.CompletenessFull,
		Purpose:       game.Purpose{Tokens: 1, CreatureTokens: 1},
		OptionalCosts: []game.AdditionalCost{effects.Buyback("{3}")},
		OnResolve:     effects.ResolveSteps(effects.CreateToken{Template: effects.TokenCard("1/1 white Soldier"), N: 1}),
	})
}
