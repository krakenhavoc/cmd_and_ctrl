package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	Register(Spec{
		OracleID:      "e25e2a3f-4add-4744-b1f8-2e880998a51f",
		Name:          "Recruit the Worthy",
		Completeness:  CompletenessFull,
		Purpose:       game.Purpose{Tokens: 1, CreatureTokens: 1},
		OptionalCosts: []game.AdditionalCost{Buyback("{3}")},
		OnResolve:     ResolveSteps(CreateToken{Template: TokenCard("1/1 white Soldier"), N: 1}),
	})
}
