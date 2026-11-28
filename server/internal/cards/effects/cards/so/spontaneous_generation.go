package cardsso

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Spontaneous Generation — Sorcery {3}{G} (Mercadian Masques, S82 OLD batch 697, #5619):
//
//	"Create a 1/1 green Saproling creature token for each card in your hand."
//
// The hand is counted as the spell resolves (CR 608.2h); the spell itself is
// on the stack, not in the hand.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "d5b021e3-a0c1-471b-9088-a76816806354",
		Name:         "Spontaneous Generation",
		Completeness: effects.CompletenessFull,
		OnResolve: effects.ResolveSteps(effects.CreateTokensForEach{
			Template: effects.TokenCard("1/1 green Saproling"),
			Name:     "each card in your hand",
			Count:    effects.SokB1CardsInYourHand,
		}),
	})
}
