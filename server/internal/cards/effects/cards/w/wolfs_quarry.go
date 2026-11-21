package cardsw

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Wolf's Quarry — Sorcery {4}{G}{G} (Jumpstart 2022, #5948):
//
//	"Create three 1/1 green Boar creature tokens with "When this token
//	 dies, create a Food token.""
//
// The Boar is a catalog token carrying its own death trigger
// (s82_gap_b176_helpers.go, ADR 0083), so each Boar makes its own Food for
// whoever controls it when it dies.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "c6fead73-0473-4b57-899a-9f8c5b4f896d",
		Name:         "Wolf's Quarry",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Tokens: 3},
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			return effects.CreateToken{Template: effects.BoarFoodToken(), N: 3}.Apply(ctx)
		},
	})
}
