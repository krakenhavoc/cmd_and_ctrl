package cardssy

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Talrand's Invocation — Sorcery {2}{U}{U} (Jumpstart 2020):
//
//	"Create two 2/2 blue Drake creature tokens with flying."
//
// The Drake is the same token row Talrand, Sky Summoner makes.
//
// No simplification. Deck request #4894 (Jumpstart Azorius).
func init() {
	effects.Register(effects.Spec{
		OracleID:     "43355ac4-bf8c-48f6-a322-fafbc9d132d1",
		Name:         "Talrand's Invocation",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Tokens: 2},
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			return effects.CreateToken{Controller: item.Controller, Template: effects.TokenCard("2/2 blue Drake with flying"), N: 2}.Apply(ctx)
		},
	})
}
